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
	"runtime/debug"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/cmd/config"
	server_sync_tasks "github.com/movsb/taoblog/cmd/server/tasks/sync"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/gateway/addons"
	"github.com/movsb/taoblog/gateway/handlers/rss"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/backups"
	backups_git "github.com/movsb/taoblog/modules/backups/git"
	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/metrics"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/cache"
	"github.com/movsb/taoblog/service/modules/notify"
	"github.com/movsb/taoblog/service/modules/notify/mailer"
	"github.com/movsb/taoblog/service/modules/request_throttler"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
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

func AddCommands(rootCmd *cobra.Command) {
	serveCommand := &cobra.Command{
		Use:   `server`,
		Short: `Run the server`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var cfg *config.Config
			dir := os.DirFS(`.`)
			demo := utils.Must1(cmd.Flags().GetBool(`demo`))
			if demo {
				cfg = config.DefaultDemoConfig()
				// 并且强制关闭本地环境。
				version.ForceEnableDevMode = `0`
			} else {
				cfg2 := config.DefaultConfig()
				if err := config.ApplyFromFile(cfg2, dir, `taoblog.yml`); err != nil {
					if !os.IsNotExist(err) {
						log.Fatalln(err)
					}
				}
				cfg = cfg2
			}
			configOverride := func(cfg *config.Config) {
				if err := config.ApplyFromFile(cfg, dir, `taoblog.override.yml`); err != nil {
					if !os.IsNotExist(err) {
						log.Fatalln(err)
					}
				}
			}
			s := NewServer(
				WithRequestThrottler(request_throttler.New()),
				WithCreateFirstPost(),
				WithGitSyncTask(true),
				WithBackupTasks(true),
				WithRSS(true),
				WithMonitorCerts(true),
				WithMonitorDomain(true),
				WithConfigOverride(configOverride),
			)
			s.Serve(context.Background(), false, cfg, nil)
		},
	}

	serveCommand.Flags().Bool(`demo`, false, `运行演示实例。`)

	rootCmd.AddCommand(serveCommand)
}

// 服务器实例。
type Server struct {
	testing bool

	httpAddr   string
	httpServer *http.Server

	grpcAddr string

	// 请求节流器。
	throttler        grpc.UnaryServerInterceptor
	throttlerEnabled atomic.Bool

	createFirstPost   bool
	initialTimezone   *time.Location
	initGitSyncTask   bool
	initBackupTasks   bool
	initRssTasks      bool
	initMonitorCerts  bool
	initMonitorDomain bool
	configOverride    func(cfg *config.Config)

	db      *taorm.DB
	auth    *auth.Auth
	main    *service.Service
	gateway *gateway.Gateway
	rss     *rss.RSS

	fileCache *cache.FileCache

	metrics *metrics.Registry

	notifyServer proto.NotifyServer
}

// 主要用于测试启动。
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
	if s.auth == nil {
		panic(`auth service is not created`)
	}
	return s.auth
}
func (s *Server) Main() *service.Service {
	if s.main == nil {
		panic(`main service is not created`)
	}
	return s.main
}
func (s *Server) DB() *taorm.DB {
	return s.db
}
func (s *Server) Gateway() *gateway.Gateway {
	return s.gateway
}
func (s *Server) RSS() *rss.RSS {
	if s.rss == nil {
		panic(`rss is not created`)
	}
	return s.rss
}

func (s *Server) Serve(ctx context.Context, testing bool, cfg *config.Config, ready chan<- struct{}) {
	if s.httpAddr != `` {
		panic(`server is already running`)
	}

	s.testing = testing

	log.Println(`DevMode:`, version.DevMode())
	log.Println(`Time.Now:`, time.Now().Format(time.RFC3339))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rc := runtime_config.NewRuntime()
	ctx = runtime_config.Context(ctx, rc)

	cacheDB := migration.InitCache(cfg.Database.Cache)
	s.fileCache = cache.NewFileCache(ctx, cacheDB)

	postsDB := migration.InitPosts(cfg.Database.Posts, s.createFirstPost)
	defer postsDB.Close()

	s.db = taorm.NewDB(postsDB)

	filesDB := migration.InitFiles(cfg.Database.Files)
	filesStore := theme_fs.FS(storage.NewSQLite(postsDB, storage.NewDataStore(filesDB)))

	migration.Migrate(postsDB, filesDB, cacheDB)

	utils.Must(s.initConfig(cfg, postsDB))

	s.metrics = metrics.NewRegistry(context.TODO())

	var mux = http.NewServeMux()
	mux.Handle(`/v3/metrics`, s.metrics.Handler()) // TODO: insecure

	theAuth := auth.New(taorm.NewDB(postsDB))
	s.auth = theAuth

	startGRPC, serviceRegistrar := s.serveGRPC(ctx)

	notify := s.createNotifyService(ctx, postsDB, cfg, serviceRegistrar)
	s.notifyServer = notify

	theService := s.createMainServices(ctx, postsDB, cfg, serviceRegistrar, notify, cancel, theAuth, filesStore, rc, mux)
	s.main = theService

	if testing && s.initialTimezone != nil {
		theService.TestingSetTimezone(s.initialTimezone)
	}

	go startGRPC()

	s.metrics.MustRegister(theService.Exporter())

	s.gateway = gateway.NewGateway(s.grpcAddr, theService, theAuth, mux, notify)
	s.gateway.SetFavicon(theService.Favicon())
	s.gateway.SetDynamic(theService.DropAllPostAndCommentCache)
	s.gateway.SetAvatar(s.fileCache, s.Main().ResolveAvatar)

	if s.initRssTasks {
		s.initRSS()
	}

	s.createAdmin(ctx, cfg, postsDB, theService, theAuth, mux)

	theme := theme.New(ctx, version.DevMode(), cfg, theService, theService, theService, theAuth)
	canon := canonical.New(theme, theService, s.metrics)
	mux.Handle(`/`, canon)

	s.serveHTTP(ctx, cfg.Server.HTTPListen, mux)

	s.sendNotify(`网站状态`, `现在开始运行。`)

	if !version.DevMode() {
		go liveCheck(ctx, s, theService)
	}

	if s.initBackupTasks {
		go s.createBackupTasks(ctx, cfg)
	}
	if s.initGitSyncTask {
		go s.createGitSyncTasks(ctx, clients.NewFromAddress(s.GRPCAddr(), auth.SystemToken()))
	}
	if s.initMonitorDomain {
		theService.SetDomainDays(-1)
		go monitorDomain(ctx, cfg.Site.Home, notify, cfg.Others.Whois.ApiLayer.Key, func(days int) {
			theService.SetDomainDays(days)
		})
	}
	if s.initMonitorCerts {
		theService.SetCertDays(-1)
		go monitorCert(ctx, cfg.Site.Home, notify, func(days int) {
			theService.SetCertDays(days)
		})
	}

	s.initSyncs(ctx, cfg, filesStore)

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

	cancel()
	<-ctx.Done()
}

func (s *Server) initConfig(cfg *config.Config, db *sql.DB) error {
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

	// 加载覆盖配置。
	if s.configOverride != nil {
		s.configOverride(cfg)
		log.Println(`加载覆盖配置`)
	}

	return nil
}

func (s *Server) initRSS() {
	client := clients.NewFromAddress(s.GRPCAddr(), ``)
	rss := rss.New(s.auth, client,
		rss.WithArticleCount(10),
		rss.WithCurrentLocationGetter(s.Main()),
	)
	s.gateway.SetRSS(rss)
	s.rss = rss
}

func (s *Server) createAdmin(ctx context.Context, cfg *config.Config, db *sql.DB, theService *service.Service, theAuth *auth.Auth, mux *http.ServeMux) {
	prefix := `/admin/`

	u, err := url.Parse(cfg.Site.Home)
	if err != nil {
		panic(err)
	}

	a := admin.NewAdmin(version.DevMode(), s.Gateway(), theService, theAuth, prefix, u.Hostname(), cfg.Site.Name, []string{u.String()},
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
	p := auth.NewPasskeys(
		taorm.NewDB(db), wa,
		theAuth.GenCookieForPasskeys,
		theAuth.DropUserCache,
	)
	theService.AuthServer = p
	theAuth.Passkeys = p
}

func (s *Server) sendNotify(title, message string) {
	s.notifyServer.SendInstant(
		auth.SystemForLocal(context.Background()),
		&proto.SendInstantRequest{
			Title: title,
			Body:  message,
		},
	)
}

func (s *Server) createGitSyncTasks(
	ctx context.Context,
	client *clients.ProtoClient,
) {
	gs := backups_git.New(ctx, client, false)

	sync := func() error {
		if version.DevMode() {
			log.Println(`开发模式不运行 git 同步`)
			return nil
		}

		err := gs.Sync()
		if err == nil {
			log.Println(`git 同步成功`)
			/*
				s.notifyServer.SendInstant(
					auth.SystemForLocal(context.Background()),
					&proto.SendInstantRequest{
						Title: `同步成功`,
						Body:  `全部完成，没有错误。`,
						Group: `同步与备份`,
						Level: proto.SendInstantRequest_Passive,
					},
				)
			*/
			s.main.SetLastSyncAt(time.Now())
		} else {
			s.notifyServer.SendInstant(
				auth.SystemForLocal(context.Background()),
				&proto.SendInstantRequest{
					Title: `同步失败`,
					Body:  err.Error(),
					Group: `同步与备份`,
					Level: proto.SendInstantRequest_Active,
				},
			)
		}

		return err
	}

	// log.Println(sync())

	const every = time.Hour * 1
	ticker := time.NewTicker(every)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println(`git 同步任务退出`)
			return
		case <-gs.Do():
			log.Println(`立即执行同步中`)
			if err := sync(); err != nil {
				log.Println(err)
			}
		case <-ticker.C:
			if err := sync(); err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *Server) createBackupTasks(
	ctx context.Context,
	cfg *config.Config,
) {
	client := clients.NewFromAddress(s.GRPCAddr(), auth.SystemToken())
	ctx = auth.SystemForGateway(ctx)
	if r2 := cfg.Maintenance.Backups.R2; r2.Enabled && !version.DevMode() {
		b := utils.Must1(backups.New(
			ctx, s.main.GetPluginStorage(`backups.r2`), client,
			backups.WithRemoteOSS(`r2`, &r2.OSSConfig),
			backups.WithEncoderAge(r2.AgeKey),
		))

		execute := func() {
			var messages []string
			var hasErr bool
			if err := b.BackupPosts(ctx); err != nil {
				log.Println(`备份失败：`, err)
				messages = append(messages, fmt.Sprintf(`文章备份失败：%v`, err))
				hasErr = true
			} else {
				log.Println(`文章备份成功`)
				// 通知过于频繁，改成指标，手机端通过小组件展。
				// messages = append(messages, `文章备份成功。`)
				s.main.SetLastBackupAt(time.Now())
			}
			if err := b.BackupFiles(ctx); err != nil {
				log.Println(`备份失败：`, err)
				messages = append(messages, fmt.Sprintf(`附件备份失败：%v`, err))
				hasErr = true
			} else {
				log.Println(`附件备份成功`)
				// messages = append(messages, `附件备份成功。`)
				s.main.SetLastBackupAt(time.Now())
			}
			if len(messages) > 0 {
				s.notifyServer.SendInstant(
					auth.SystemForLocal(context.Background()),
					&proto.SendInstantRequest{
						Title: `文章和附件备份`,
						Body:  strings.Join(messages, "\n"),
						Group: `同步与备份`,
						Level: utils.IIF(hasErr, proto.SendInstantRequest_Active, proto.SendInstantRequest_Passive),
					},
				)
			}
		}

		time.Sleep(time.Minute * 10)
		execute()

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			execute()
		}
	}
}

func (s *Server) createMainServices(
	ctx context.Context, db *sql.DB, cfg *config.Config, sr grpc.ServiceRegistrar,
	notifier proto.NotifyServer,
	cancel func(),
	auth *auth.Auth,
	filesStore theme_fs.FS,
	rc *runtime_config.Runtime,
	mux *http.ServeMux,
) *service.Service {
	serviceOptions := []service.With{
		service.WithPostDataFileSystem(filesStore),
		service.WithNotifier(notifier),
		service.WithCancel(cancel),
		service.WithFileCache(s.fileCache),
	}

	addons.New()

	return service.New(ctx, sr, cfg, db, rc, auth, mux, serviceOptions...)
}

func (s *Server) createNotifyService(ctx context.Context, db *sql.DB, cfg *config.Config, sr grpc.ServiceRegistrar) proto.NotifyServer {
	var options []notify.With

	if ch := cfg.Notify.Bark; ch.Token != `` {
		options = append(options, notify.WithDefaultToken(ch.Token))
	}

	// TODO 移动到内部实现。
	if m := cfg.Notify.Mailer; m.Account != `` && m.Server != `` {
		mail := mailer.NewMailer(m.Server, m.Account, m.Password)
		stored := mailer.NewMailerLogger(logs.NewLogStore(db), mail)
		options = append(options, notify.WithMailer(stored))
	}

	n := notify.New(ctx, sr, db, options...)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute * 10):
				store := logs.NewLogStore(db)
				if c := store.CountStaleLogs(time.Minute * 10); c > 0 {
					n.SendInstant(auth.SystemForLocal(ctx), &proto.SendInstantRequest{
						Title:       `有堆积的日志未处理`,
						Body:        fmt.Sprintf(`条数：%d`, c),
						Immediately: true,
					})
				}
			}
		}
	}()

	return n
}

// 因为 GRPC 服务启动后不能注册，所以返回了一个函数用于适时启动。
func (s *Server) serveGRPC(ctx context.Context) (func(), grpc.ServiceRegistrar) {
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

	s.grpcAddr = l.Addr().String()
	log.Println(`GRPC:`, s.grpcAddr)

	go func() {
		<-ctx.Done()
		server.GracefulStop()
	}()

	return func() {
		defer l.Close()
		server.Serve(l)
	}, server
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

	stack := debug.Stack()
	log.Println("未处理的内部错误：", e, "\n", string(stack))
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
func liveCheck(ctx context.Context, s *Server, svc *service.Service) {
	t := time.NewTicker(time.Minute * 1)
	defer t.Stop()

	for range t.C {
		for !func() bool {
			now := time.Now()
			svc.GetPost(auth.SystemForLocal(context.TODO()), &proto.GetPostRequest{Id: 1})
			if elapsed := time.Since(now); elapsed > time.Second*10 {
				svc.MaintenanceMode().Enter(`我也不知道为什么，反正就是服务接口卡住了🥵。`, -1)
				log.Println(`服务接口响应非常慢了。`)
				s.sendNotify(`服务不可用`, `保活检测卡住了`)
				return false
			}
			svc.MaintenanceMode().Leave()
			return true
		}() {
		}
	}
}

func (s *Server) initSyncs(ctx context.Context, cfg *config.Config, filesStore theme_fs.FS) {
	// if version.DevMode() {
	// 	log.Println(`测试环境不启动同步。`)
	// 	return
	// }
	for _, backend := range []struct {
		config *config.OSSConfigWithEnabled
		name   string
		store  string
	}{
		{
			config: &cfg.Site.Sync.Aliyun,
			name:   `aliyun`,
			store:  `site.sync.aliyun`,
		},
		{
			config: &cfg.Site.Sync.COS,
			name:   `cos`,
			store:  `site.sync.cos`,
		},
		{
			config: &cfg.Site.Sync.R2,
			name:   `r2`,
			store:  `site.sync.r2`,
		},
		{
			config: &cfg.Site.Sync.Minio,
			name:   `minio`,
			store:  `site.sync.minio`,
		},
	} {
		if backend.config.Enabled {
			oss, err := server_sync_tasks.NewSyncToOSS(
				ctx, backend.name, &backend.config.OSSConfig, s.Main(),
				s.Main().GetPluginStorage(backend.store),
				filesStore,
			)
			if err != nil {
				s.sendNotify(`启动同步失败`, err.Error())
				log.Println(err)
				continue
			}
			s.Main().RegisterFileURLGetter(backend.name, oss)
			log.Println(`启动同步：`, backend.name)
		}
	}
}
