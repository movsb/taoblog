package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/movsb/pkg/notify"
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/metrics"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/theme"
	"github.com/movsb/taoblog/theme/modules/canonical"
	"github.com/spf13/cobra"
)

// AddCommands ...
func AddCommands(rootCmd *cobra.Command) {
	serveCommand := &cobra.Command{
		Use:   `server`,
		Short: `Run the server`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			serve()
		},
	}

	rootCmd.AddCommand(serveCommand)
}

func serve() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cfg := config.LoadFile(`taoblog.yml`)

	db := initDatabase(cfg)
	defer db.Close()

	migration.Migrate(db)

	log.Println(`DevMode:`, service.DevMode())

	var mux = http.NewServeMux()
	r := metrics.NewRegistry(context.TODO())
	mux.Handle(`/v3/metrics`, r.Handler()) // TODO: insecure

	theAuth := auth.New(cfg.Auth, service.DevMode())
	theService := service.NewService(cfg, db, theAuth)
	theAuth.SetService(theService)

	theAuth.SetAdminWebAuthnCredentials(theService.GetDefaultStringOption(`admin_webauthn_credentials`, "[]"))

	gateway.NewGateway(theService, theAuth, mux)

	if !cfg.Maintenance.DisableAdmin {
		prefix := `/admin/`

		u, err := url.Parse(cfg.Site.Home)
		if err != nil {
			panic(err)
		}

		a := admin.NewAdmin(service.DevMode(), theService, theAuth, prefix, u.Hostname(), cfg.Site.Name, []string{u.String()})
		log.Println(`admin on`, prefix)
		mux.Handle(prefix, a.Handler())
	}

	theme := theme.New(service.DevMode(), cfg, theService, theAuth)
	canon := canonical.New(theme, r)
	mux.Handle(`/`, canon)

	server := &http.Server{
		Addr: cfg.Server.HTTPListen,
		Handler: utils.ChainFuncs(
			http.Handler(mux),
			theAuth.UserFromCookieHandler,
			logs.NewRequestLoggerHandler(`access.log`),
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

	var chanify *notify.Chanify
	log.Println("Server started on", server.Addr)
	if cc := cfg.Comment.Push.Chanify; cc.Token != "" {
		if cfg.Comment.Notify {
			chanify = notify.NewOfficialChanify(cc.Token)
			chanify.Send("博客状态", "已经开始运行。", true)
		}
	}

	go liveCheck(theService, chanify)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit
	close(quit)

	log.Println("server shutting down")
	theService.MaintenanceMode().Enter(`服务重启中...`, time.Second*30)
	server.Shutdown(context.Background())
	log.Println("server shut down")
}

// TODO 文章 1 必须存在。可以是非公开状态。
func liveCheck(s *service.Service, cc *notify.Chanify) {
	t := time.NewTicker(time.Minute * 15)
	defer t.Stop()

	for range t.C {
		for !func() bool {
			now := time.Now()
			s.GetPost(context.Background(), &protocols.GetPostRequest{Id: 1})
			if elapsed := time.Since(now); elapsed > time.Second*10 {
				s.MaintenanceMode().Enter(`我也不知道为什么，反正就是服务接口卡住了🥵。`, -1)
				log.Println(`服务接口响应非常慢了。`)
				if cc != nil {
					cc.Send(`服务不可用`, `保活检测卡住了。`, true)
				}
				return false
			}
			s.MaintenanceMode().Leave()
			return true
		}() {
		}
	}
}

func initDatabase(cfg *config.Config) *sql.DB {
	var db *sql.DB
	var err error

	switch cfg.Database.Engine {
	case `sqlite`:
		v := url.Values{}
		v.Set(`cache`, `shared`)
		v.Set(`mode`, `rwc`)
		u := url.URL{
			Scheme:   `file`,
			Opaque:   url.PathEscape(cfg.Database.SQLite.Path),
			RawQuery: v.Encode(),
		}
		db, err = sql.Open(`sqlite3`, u.String())
		if err == nil {
			db.SetMaxOpenConns(1)
		}
	default:
		panic(`unknown database engine`)
	}
	if err != nil {
		panic(err)
	}

	var count int
	row := db.QueryRow(`select count(1) from options`)
	if err := row.Scan(&count); err != nil {
		if se, ok := err.(sqlite3.Error); ok {
			if strings.Contains(se.Error(), `no such table`) {
				migration.Init(cfg, db)
				return db
			}
		}
		panic(err)
	}
	return db
}
