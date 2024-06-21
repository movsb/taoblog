package server

import (
	"context"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/mattn/go-sqlite3"
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/metrics"
	"github.com/movsb/taoblog/modules/metrics/exporters/hostdare"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/theme"
	"github.com/movsb/taoblog/theme/modules/canonical"
	"github.com/movsb/taorm"
	"github.com/spf13/cobra"
)

// AddCommands ...
func AddCommands(rootCmd *cobra.Command) {
	serveCommand := &cobra.Command{
		Use:   `server`,
		Short: `Run the server`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			serve(context.Background())
		},
	}

	rootCmd.AddCommand(serveCommand)
}

func serve(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cfg := config.LoadFile(`taoblog.yml`)

	db := InitDatabase(cfg.Database.Path)
	defer db.Close()

	migration.Migrate(db)

	// 在此之前不能读配置！！！
	updater := config.NewUpdater(cfg)
	updater.EachSaver(func(path string, obj any) {
		// TODO 改成 grpc 配置服务。
		var option models.Option
		err := taorm.NewDB(db).Model(option).Where(`name=?`, path).Find(&option)
		if err != nil {
			if !taorm.IsNotFoundError(err) {
				panic(err)
			}
			return
		}
		if err := json.Unmarshal([]byte(option.Value), obj); err != nil {
			panic(err)
		}
		log.Println(`加载配置：`, path)
	})

	log.Println(`DevMode:`, service.DevMode())

	instantNotifier := notify.NewConsoleNotify()
	if token := cfg.Notify.Chanify.Token; token != "" {
		instantNotifier = notify.NewChanifyInstantNotify(token)
	}

	var mux = http.NewServeMux()
	r := metrics.NewRegistry(context.TODO())
	mux.Handle(`/v3/metrics`, r.Handler()) // TODO: insecure
	if hd := cfg.VPS.Hostdare; hd.Username != "" && hd.Password != "" {
		exporter, err := hostdare.New(hd.Username, hd.Password)
		if err != nil {
			log.Println(err)
		} else {
			r.MustRegister(exporter)
			log.Println(`注册 hostdare 指标`)
		}
	}

	storage := storage.NewLocal(cfg.Data.File.Path)

	theAuth := auth.New(cfg.Auth, service.DevMode())
	theService := service.NewService(ctx, cancel, cfg, db, theAuth,
		service.WithPostDataFileSystem(storage),
		service.WithInstantNotifier(instantNotifier),
	)
	r.MustRegister(theService.Exporter())
	theAuth.SetService(theService)

	theAuth.SetAdminWebAuthnCredentials(theService.GetDefaultStringOption(`admin_webauthn_credentials`, "[]"))

	gateway.NewGateway(theService, theAuth, mux, instantNotifier)

	if !cfg.Maintenance.DisableAdmin {
		prefix := `/admin/`

		u, err := url.Parse(cfg.Site.Home)
		if err != nil {
			panic(err)
		}

		a := admin.NewAdmin(service.DevMode(), theService, theAuth, prefix, u.Hostname(), cfg.Site.Name, []string{u.String()},
			admin.WithCustomThemes(&cfg.Theme),
		)
		log.Println(`admin on`, prefix)
		mux.Handle(prefix, a.Handler())

		config := &webauthn.Config{
			RPID:          u.Hostname(),
			RPDisplayName: cfg.Site.Name,
			RPOrigins:     []string{u.String()},
		}
		wa, err := webauthn.New(config)
		if err != nil {
			panic(err)
		}
		theService.AuthServer = auth.NewPasskeys(
			wa,
			func(userHandler []byte) (*auth.User, string, error) {
				id := binary.LittleEndian.Uint32(userHandler)
				u := theAuth.GetUserByID(int64(id))
				if u.IsAdmin() {
					return u, fmt.Sprintf("%d:%s", u.ID, cfg.Auth.Key), nil
				}

				return u, "", errors.New(`no such user`)
			},
			theAuth.GenCookieForPasskeys,
		)
	}

	theme := theme.New(service.DevMode(), cfg, theService, theService, theService, theAuth, storage)
	canon := canonical.New(theme, r)
	mux.Handle(`/`, canon)

	server := &http.Server{
		Addr: cfg.Server.HTTPListen,
		Handler: utils.ChainFuncs(
			http.Handler(mux),
			// 注意这个拦截器的能力：
			//
			// 所有进入服务端认证信息均被包含在 context 中，
			// 这也包含了 Gateway。
			//
			// 但是，gateway 虽然有了 auth context，但是如果使用的是 grpc-client，
			// 无法传递给 server，会再次用 auth.NewContextForRequestAsGateway 再度解析并传递。
			theAuth.UserFromCookieHandler,
			logs.NewRequestLoggerHandler(`access.log`, logs.WithSentBytesCounter(r)),
			theService.MaintenanceMode().Handler(func(ctx context.Context) bool {
				return auth.Context(ctx).User.IsAdmin()
			}),
		),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()

	log.Println("Server started on", server.Addr)
	instantNotifier.InstantNotify("博客状态", "已经开始运行。")

	go liveCheck(theService, instantNotifier)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)

	select {
	case <-quit:
	case <-ctx.Done():
	}

	log.Println("server shutting down")
	theService.MaintenanceMode().Enter(`服务关闭中...`, time.Second*30)
	server.Shutdown(context.Background())
	log.Println("server shut down")
}

// TODO 文章 1 必须存在。可以是非公开状态。
// TODO 放在服务里面 tasks.go
func liveCheck(s *service.Service, cc notify.InstantNotifier) {
	t := time.NewTicker(time.Minute * 15)
	defer t.Stop()

	for range t.C {
		for !func() bool {
			now := time.Now()
			s.GetPost(auth.SystemAdmin(context.TODO()), &proto.GetPostRequest{Id: 1})
			if elapsed := time.Since(now); elapsed > time.Second*10 {
				s.MaintenanceMode().Enter(`我也不知道为什么，反正就是服务接口卡住了🥵。`, -1)
				log.Println(`服务接口响应非常慢了。`)
				if cc != nil {
					cc.InstantNotify(`服务不可用`, `保活检测卡住了。`)
				}
				return false
			}
			s.MaintenanceMode().Leave()
			return true
		}() {
		}
	}
}

func InitDatabase(path string) *sql.DB {
	var db *sql.DB
	var err error

	v := url.Values{}
	v.Set(`cache`, `shared`)
	v.Set(`mode`, `rwc`)

	if path == `` {
		// 内存数据库
		path = `no-matter-what-path-used`
		v.Set(`mode`, `memory`)
	}

	u := url.URL{
		Scheme:   `file`,
		Opaque:   url.PathEscape(path),
		RawQuery: v.Encode(),
	}

	db, err = sql.Open(`sqlite3`, u.String())
	if err == nil {
		db.SetMaxOpenConns(1)
	}
	if err != nil {
		panic(err)
	}

	var count int
	row := db.QueryRow(`select count(1) from options`)
	if err := row.Scan(&count); err != nil {
		if se, ok := err.(sqlite3.Error); ok {
			if strings.Contains(se.Error(), `no such table`) {
				migration.Init(db, path)
				return db
			}
		}
		panic(err)
	}
	return db
}
