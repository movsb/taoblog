package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/mattn/go-sqlite3"
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/mailer"
	"github.com/movsb/taoblog/modules/metrics"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/request_throttler"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/theme"
	"github.com/movsb/taoblog/theme/modules/canonical"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/movsb/taorm"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AddCommands ...
func AddCommands(rootCmd *cobra.Command) {
	serveCommand := &cobra.Command{
		Use:   `server`,
		Short: `Run the server`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.LoadFile(`taoblog.yml`)
			s := NewDefaultServer()
			s.Serve(context.Background(), false, cfg, nil)
		},
	}

	rootCmd.AddCommand(serveCommand)
}

// 服务器实例。
type Server struct {
	httpAddr   string
	httpServer *http.Server

	grpcAddr   string
	grpcServer *grpc.Server

	// 请求节流器。
	throttler        grpc.UnaryServerInterceptor
	throttlerEnabled atomic.Bool

	createFirstPost bool

	db      *taorm.DB
	auth    *auth.Auth
	main    *service.Service
	gateway *gateway.Gateway

	metrics *metrics.Registry
}

func NewDefaultServer() *Server {
	return NewServer(
		WithRequestThrottler(request_throttler.New()),
		WithCreateFirstPost(),
	)
}

func NewServer(with ...With) *Server {
	s := &Server{}
	for _, w := range with {
		w(s)
	}
	return s
}

// 运行时的真实 HTTP 地址。
// 形如：127.0.0.1:2564，不包含协议、路径等。
func (s *Server) HTTPAddr() string {
	if s.httpAddr == `` {
		panic(`no http addr`)
	}
	return s.httpAddr
}

// 运行时的真实 GRPC 地址。
// 形如：127.0.0.1:2563，不包含协议、路径等。
func (s *Server) GRPCAddr() string {
	if s.grpcAddr == `` {
		panic(`no grpc addr`)
	}
	return s.grpcAddr
}

func (s *Server) Auth() *auth.Auth {
	return s.auth
}
func (s *Server) Main() *service.Service {
	return s.main
}
func (s *Server) DB() *taorm.DB {
	return s.db
}
func (s *Server) Gateway() *gateway.Gateway {
	return s.gateway
}

func (s *Server) Serve(ctx context.Context, testing bool, cfg *config.Config, ready chan<- struct{}) {
	if s.httpAddr != `` {
		panic(`server is already running`)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	db := InitDatabase(cfg.Database.Path, s.createFirstPost)
	defer db.Close()

	s.db = taorm.NewDB(db)

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

	log.Println(`DevMode:`, version.DevMode())
	log.Println(`Time.Now:`, time.Now().Format(time.RFC3339))

	s.metrics = metrics.NewRegistry(context.TODO())

	var mux = http.NewServeMux()
	mux.Handle(`/v3/metrics`, s.metrics.Handler()) // TODO: insecure

	var store theme_fs.FS
	if path := cfg.Data.File.Path; path == "" {
		store = storage.NewMemory()
	} else {
		store = storage.NewLocal(cfg.Data.File.Path)
	}

	theAuth := auth.New(cfg.Auth, taorm.NewDB(db))
	s.auth = theAuth

	notifier := notify.NewNotifyLogger(notify.NewLogStore(db))
	if token := cfg.Notify.Chanify.Token; token != "" {
		notifier.SetNotifier(notify.NewChanifyNotify(token))
	} else {
		notifier.SetNotifier(notify.NewConsoleNotify())
	}

	mailer := mailer.NewMailerLogger(notify.NewLogStore(db), mailer.NewMailerConfig(
		cfg.Notify.Mailer.Server, cfg.Notify.Mailer.Account, cfg.Notify.Mailer.Password,
	))

	startGRPC := s.serveGRPC(ctx)

	serviceOptions := []service.With{
		service.WithPostDataFileSystem(store),
		service.WithNotifier(notifier),
		service.WithMailer(mailer),
	}

	theService := service.NewService(ctx, s.grpcServer, cancel, testing, cfg, db, theAuth, serviceOptions...)
	s.main = theService
	go startGRPC()

	s.metrics.MustRegister(theService.Exporter())

	s.gateway = gateway.NewGateway(s.grpcAddr, theService, theAuth, mux, notifier)

	(func() {
		prefix := `/admin/`

		u, err := url.Parse(cfg.Site.Home)
		if err != nil {
			panic(err)
		}

		a := admin.NewAdmin(version.DevMode(), theService, theAuth, prefix, u.Hostname(), cfg.Site.Name, []string{u.String()},
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
			taorm.NewDB(db), wa,
			theAuth.GenCookieForPasskeys,
		)
	})()

	theme := theme.New(version.DevMode(), cfg, theService, theService, theService, theAuth, store)
	canon := canonical.New(theme, s.metrics)
	mux.Handle(`/`, canon)

	s.serveHTTP(ctx, cfg.Server.HTTPListen, mux)

	notifier.Notify("博客状态", "已经开始运行。")

	go liveCheck(theService, notifier)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM)

	if ready != nil {
		ready <- struct{}{}
	}

	select {
	case <-quit:
	case <-ctx.Done():
	}

	log.Println("server shutting down")
	theService.MaintenanceMode().Enter(`服务关闭中...`, time.Second*30)
	s.httpServer.Shutdown(context.Background())
	log.Println("server shut down")
}

// 因为 GRPC 服务启动后不能注册，所以返回了一个函数用于适时启动。
//
// TODO
func (s *Server) serveGRPC(ctx context.Context) func() {
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(100<<20),
		grpc.MaxSendMsgSize(100<<20),
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(exceptionRecoveryHandler)),
			s.auth.UserFromGatewayUnaryInterceptor(),
			s.auth.UserFromClientTokenUnaryInterceptor(),
			s.throttlerGatewayInterceptor,
			grpcLoggerUnary,
		),
		grpc_middleware.WithStreamServerChain(
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(exceptionRecoveryHandler)),
			s.auth.UserFromGatewayStreamInterceptor(),
			s.auth.UserFromClientTokenStreamInterceptor(),
			grpcLoggerStream,
		),
	)

	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", `127.0.0.1:0`)
	if err != nil {
		panic(err)
	}
	s.grpcServer = server
	s.grpcAddr = l.Addr().String()
	log.Println(`GRPC listen on:`, s.grpcAddr)

	return func() {
		defer l.Close()
		server.Serve(l)
	}
}

func (s *Server) throttlerGatewayInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if s.throttlerEnabled.Load() && s.throttler != nil {
		return s.throttler(ctx, req, info, handler)
	}
	return handler(ctx, req)
}

// 运行 HTTP 服务。
// 真实地址回写到 s.httpAddr 中。
func (s *Server) serveHTTP(ctx context.Context, addr string, h http.Handler) {
	server := &http.Server{
		Addr: addr,
		Handler: utils.ChainFuncs(h,
			// 注意这个拦截器的能力：
			//
			// 所有进入服务端认证信息均被包含在 context 中，
			// 这也包含了 Gateway。
			//
			// 但是，gateway 虽然有了 auth context，但是如果使用的是 grpc-client，
			// 无法传递给 server，会再次用 auth.NewContextForRequestAsGateway 再度解析并传递。
			s.auth.UserFromCookieHandler,
			logs.NewRequestLoggerHandler(`access.log`, logs.WithSentBytesCounter(s.metrics)),
			s.main.MaintenanceMode().Handler(func(ctx context.Context) bool {
				return auth.Context(ctx).User.IsAdmin()
			}),
		),
	}

	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", server.Addr)
	if err != nil {
		panic(err)
	}
	// 总是会被 http.Server 关闭。
	// defer l.Close()
	s.httpAddr = l.Addr().String()
	s.httpServer = server
	log.Println(`HTTP:`, s.httpAddr)

	go func() {
		if err := server.Serve(l); err != nil {
			if err != http.ErrServerClosed {
				panic(err)
			}
		}
	}()
}

func exceptionRecoveryHandler(e any) error {
	switch te := e.(type) {
	case *taorm.Error:
		switch typed := te.Err.(type) {
		case *taorm.NotFoundError:
			return status.New(codes.NotFound, typed.Error()).Err()
		case *taorm.DupKeyError:
			return status.New(codes.AlreadyExists, typed.Error()).Err()
		}
	case *status.Status:
		return te.Err()
	case codes.Code:
		return status.Error(te, te.String())
	case error:
		if st, ok := status.FromError(te); ok {
			return st.Err()
		}
	}
	buf := make([]byte, 10<<10)
	runtime.Stack(buf, false)
	log.Println("未处理的内部错误：", e, "\n", string(buf))
	return status.New(codes.Internal, fmt.Sprint(e)).Err()
}

func grpcLoggerUnary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	grpcLogger(ctx, info.FullMethod)
	return handler(ctx, req)
}
func grpcLoggerStream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	grpcLogger(ss.Context(), info.FullMethod)
	return handler(srv, ss)
}

var enableGrpcLogger = expvar.NewInt(`log.grpc`)

func grpcLogger(ctx context.Context, method string) {
	logEnabled := enableGrpcLogger.Value() == 1
	if !logEnabled {
		return
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Println(md)
	}
	ac := auth.Context(ctx)
	log.Println(method, ac.UserAgent)
}

// TODO 文章 1 必须存在。可以是非公开状态。
// TODO 放在服务里面 tasks.go
// TODO 放在 daemon 里面（同 webhooks）
func liveCheck(s *service.Service, cc notify.Notifier) {
	t := time.NewTicker(time.Minute * 1)
	defer t.Stop()

	for range t.C {
		for !func() bool {
			now := time.Now()
			s.GetPost(auth.SystemAdmin(context.TODO()), &proto.GetPostRequest{Id: 1})
			if elapsed := time.Since(now); elapsed > time.Second*10 {
				s.MaintenanceMode().Enter(`我也不知道为什么，反正就是服务接口卡住了🥵。`, -1)
				log.Println(`服务接口响应非常慢了。`)
				if cc != nil {
					cc.Notify(`服务不可用`, `保活检测卡住了。`)
				}
				return false
			}
			s.MaintenanceMode().Leave()
			return true
		}() {
		}
	}
}

// 如果路径为空，使用内存数据库。
// TODO 把首篇文章的创建改到服务中实现，属于业务逻辑。
func InitDatabase(path string, createFirstPost bool) *sql.DB {
	var db *sql.DB
	var err error

	v := url.Values{}
	v.Set(`cache`, `shared`)
	v.Set(`mode`, `rwc`)

	if path == `` {
		// 内存数据库
		// NOTE: 测试的时候同名路径会引用同一个内存数据库，
		// 所以需要取不同的路径名。
		path = fmt.Sprintf(`%s@%d`,
			`no-matter-what-path-used`,
			time.Now().UnixNano(),
		)
		v.Set(`mode`, `memory`)
	}

	u := url.URL{
		Scheme:   `file`,
		Opaque:   url.PathEscape(path),
		RawQuery: v.Encode(),
	}

	dsn := u.String()
	log.Println(`数据库连接字符串：`, dsn)
	db, err = sql.Open(`sqlite3`, dsn)
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

				tdb := taorm.NewDB(db)
				now := time.Now().Unix()

				if createFirstPost {
					tdb.MustTxCall(func(tx *taorm.DB) {
						tx.Model(&models.Post{
							Date:       int32(now),
							Modified:   int32(now),
							Title:      `你好，世界`,
							Type:       `post`,
							Category:   0,
							Status:     `public`,
							SourceType: `markdown`,
							Source:     `你好，世界！这是您的第一篇文章。`,

							// TODO 用配置时区。
							DateTimezone: ``,
							// TODO 用配置时区。
							ModifiedTimezone: ``,
						}).MustCreate()
					})
				}

				return db
			}
		}
		panic(err)
	}
	return db
}
