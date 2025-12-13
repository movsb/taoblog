package server

import (
	"context"
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/movsb/taoblog/admin"
	"github.com/movsb/taoblog/cmd/config"
	server_auth "github.com/movsb/taoblog/cmd/server/auth"
	"github.com/movsb/taoblog/cmd/server/tasks/expiration"
	"github.com/movsb/taoblog/cmd/server/tasks/git_repo"
	"github.com/movsb/taoblog/cmd/server/tasks/live_check"
	"github.com/movsb/taoblog/cmd/server/tasks/sync_files"
	"github.com/movsb/taoblog/cmd/server/tasks/year_progress"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/gateway/handlers/rss"
	"github.com/movsb/taoblog/modules/backups"
	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/micros/auth"
	"github.com/movsb/taoblog/service/micros/auth/client_login"
	"github.com/movsb/taoblog/service/micros/auth/passkeys"
	"github.com/movsb/taoblog/service/micros/auth/user"
	micros_utils "github.com/movsb/taoblog/service/micros/utils"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/cache"
	"github.com/movsb/taoblog/service/modules/notify"
	"github.com/movsb/taoblog/service/modules/notify/mailer"
	"github.com/movsb/taoblog/service/modules/renderers/auto_image_border"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taoblog/theme"
	"github.com/movsb/taoblog/theme/modules/canonical"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/movsb/taorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// 服务器实例。
type Server struct {
	testing bool

	// 运行时的真实 HTTP 地址。
	// 形如：127.0.0.1:2564，不包含协议、路径等。
	httpAddr   string
	httpServer *http.Server

	grpcAddr string

	// 请求节流器。
	throttler        grpc.UnaryServerInterceptor
	throttlerEnabled atomic.Bool

	createFirstPost  bool
	initialTimezone  *time.Location
	initGitSyncTask  bool
	initBackupTasks  bool
	initRssTasks     bool
	initMonitorCerts bool
	initYearProgress bool
	initLiveCheck    bool

	initMonitorDomain      bool
	initMonitorDomainDelay bool

	configOverride func(cfg *config.Config)

	authFrontend       *auth.Auth
	userManager        *auth.UserManager
	clientLoginService *client_login.ClientLoginService

	// 由于是先启动 grpc 才会提供给各服务注册，所以中间件这里
	// 暂时拿不到底层的 auth 接口，真正启动 grpc 时才设置。
	authMiddleware *server_auth.Middleware

	main    *service.Service
	gateway *gateway.Gateway
	rss     *rss.RSS
	utils   *micros_utils.Utils

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

// 基于服务器运行时的真实 HTTP 地址创建请求路径。
//
// NOTE：不同于 path.Join，JoinPath 会保留最后的 /（如果有的话）。
func (s *Server) JoinPath(paths ...string) string {
	p := utils.Must1(url.Parse(`http://` + s.httpAddr))
	return p.JoinPath(paths...).String()
}

// 运行时的真实 GRPC 地址。
// 形如：127.0.0.1:2563，不包含协议、路径等。
func (s *Server) GRPCAddr() string {
	if s.grpcAddr == `` {
		panic(`no grpc addr`)
	}
	return s.grpcAddr
}

func (s *Server) Auth() *auth.UserManager {
	if s.userManager == nil {
		panic(`auth service is not created`)
	}
	return s.userManager
}
func (s *Server) AuthFrontend() *auth.Auth {
	if s.authFrontend == nil {
		panic(`auth frontend is empty`)
	}
	return s.authFrontend
}
func (s *Server) Main() *service.Service {
	if s.main == nil {
		panic(`main service is not created`)
	}
	return s.main
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rc := runtime_config.NewRuntime()
	ctx = runtime_config.Context(ctx, rc)

	postsDB, _, cacheDB, filesStore, dbStatsFunc, closeAllDatabases := s.initDatabases(ctx, cfg)
	defer closeAllDatabases()

	utils.Must(s.initConfigFromDatabase(cfg, postsDB))

	log.Println(`DevMode:`, version.DevMode())
	log.Println(`Time.Now:`, time.Now().Format(time.RFC3339))
	log.Println(`Home:`, cfg.Site.GetHome())

	s.authMiddleware = server_auth.NewMiddleware()

	mux := http.NewServeMux()
	startGRPC, serviceRegistrar := s.serveGRPC(ctx)

	s.createUtilsService(ctx, cfg, serviceRegistrar)
	s.createNotifyService(ctx, postsDB, cfg, serviceRegistrar)
	s.createAuthServices(ctx, cfg, serviceRegistrar, postsDB)

	fileCache := cache.NewFileCache(ctx, cacheDB)
	s.createMainServices(ctx, postsDB, cfg, serviceRegistrar, cancel, filesStore, fileCache, rc, mux)

	// 虽然是异步开始的，但是内部只是开始 Accept 连接。
	// 如果有接口在这之前就发生了调用，也不会出问题（在 backlog 中）。
	go startGRPC()

	go dbStatsFunc(func(posts, files int64) {
		s.Main().SetPostsStorageSize(posts)
		s.Main().SetFilesStorageSize(files)
	})

	if testing && s.initialTimezone != nil {
		s.Main().TestingSetTimezone(s.initialTimezone)
	}

	s.createGateway(ctx, mux, fileCache)
	s.createAdmin(ctx, cfg, s.Main(), mux)
	s.createTheme(ctx, cfg, mux)

	s.serveHTTP(ctx, cfg.Server.HTTPListen, mux)

	s.initSubTasks(ctx, cfg, filesStore)

	if !version.DevMode() {
		s.sendNotify(`网站状态`, `现在开始运行。`)
	}

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
	s.Main().MaintenanceMode().Enter(`服务关闭中...`, time.Second*30)
	s.httpServer.Shutdown(context.Background())
	log.Println("server shut down")

	cancel()
	<-ctx.Done()
}

func (s *Server) initConfigFromDatabase(cfg *config.Config, db *taorm.DB) error {
	updater := config.NewUpdater(cfg)
	updater.EachSaver(func(path string, obj any) {
		// TODO 改成 grpc 配置服务。
		var option models.Option
		err := db.Model(option).Where(`name=?`, path).Find(&option)
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

	var options []*models.Option
	db.Where(`name LIKE 'config:%'`).MustFind(&options)
	for _, opt := range options {
		path := strings.TrimPrefix(opt.Name, `config:`)
		updater.MustApply(path, opt.Value, func(path, value string) {
			log.Println(`加载配置：`, path)
		})
	}

	// 加载覆盖配置。
	if s.configOverride != nil {
		s.configOverride(cfg)
		log.Println(`加载覆盖配置`)
	}

	return nil
}

func (s *Server) initDatabases(ctx context.Context, cfg *config.Config) (
	postsDB, filesDB, cacheDB *taorm.DB,
	filesStore *storage.SQLite,
	statsFunc func(func(posts int64, files int64)),
	closer func(),
) {
	postsDBRaw := migration.InitPosts(cfg.Database.Posts, false, s.createFirstPost)
	filesDBRaw := migration.InitFiles(cfg.Database.Files)
	cacheDBRaw := migration.InitCache(cfg.Database.Cache)

	migration.Migrate(postsDBRaw, filesDBRaw, cacheDBRaw)

	postsDB = taorm.NewDB(postsDBRaw)
	cacheDB = taorm.NewDB(cacheDBRaw)
	filesDB = taorm.NewDB(filesDBRaw)

	tmpDir := filepath.Join(os.TempDir(), version.NameLowercase)
	os.Mkdir(tmpDir, 0755)
	tmpRoot := utils.Must1(os.OpenRoot(tmpDir))
	log.Println(`临时目录：`, tmpDir)

	dataStore := storage.NewDataStore(filesDB)
	filesStore = storage.NewSQLite(postsDB, dataStore, tmpRoot)

	statsFunc = func(f func(int64, int64)) {
		stat := func() {
			var postsSize, filesSize int64
			if p := cfg.Database.Posts; p != `` {
				info, err := os.Stat(p)
				if err != nil {
					log.Println(err)
				} else {
					postsSize = info.Size()
				}
			}
			if p := cfg.Database.Files; p != `` {
				info, err := os.Stat(p)
				if err != nil {
					log.Println(err)
				} else {
					filesSize = info.Size()
				}
			}
			f(postsSize, filesSize)
		}

		stat()

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stat()
			}
		}
	}

	closer = func() {
		postsDBRaw.Close()
		filesDBRaw.Close()
		cacheDBRaw.Close()
		tmpRoot.Close()
	}

	return
}

func (s *Server) initRSS() {
	client := clients.NewFromAddress(s.GRPCAddr(), ``)
	rss := rss.New(client,
		rss.WithArticleCount(10),
		rss.WithCurrentLocationGetter(s.Main()),
	)
	s.gateway.SetRSS(rss)
	s.rss = rss
}

func (s *Server) createGateway(ctx context.Context, mux *http.ServeMux, fileCache *cache.FileCache) {
	s.gateway = gateway.NewGateway(s.grpcAddr, s.Main(), mux, s.notifyServer)
	s.gateway.SetFavicon(s.Main().Favicon())
	s.gateway.SetDynamic(s.Main().DropAllPostAndCommentCache)
	s.gateway.SetAvatar(ctx, fileCache, s.Main().ResolveAvatar)
}

func (s *Server) createAdmin(ctx context.Context, cfg *config.Config, theService *service.Service, mux *http.ServeMux) {
	prefix := `/admin/`

	a := admin.NewAdmin(
		s.gateway, theService, theService, s.Auth(),
		s.authFrontend, s.clientLoginService,
		prefix, cfg.Site.GetHome, cfg.Site.GetName,
		admin.WithCustomThemes(&cfg.Theme),
		admin.WithWebAuthnHandler(s.authFrontend.GetWebAuthnHandler),
	)

	mux.Handle(prefix, a.Handler())
}

func (s *Server) createTheme(ctx context.Context, cfg *config.Config, mux *http.ServeMux) {
	theme := theme.New(ctx, version.DevMode(), cfg, s.Main(), s.Main(), s.Main(), s.Auth(), s.authFrontend)
	canon := canonical.New(theme, s.Main())
	mux.Handle(`/`, canon)
}

func (s *Server) sendNotify(title, message string) {
	s.notifyServer.SendInstant(
		user.SystemForLocal(context.Background()),
		&proto.SendInstantRequest{
			Title: title,
			Body:  message,
		},
	)
}

func (s *Server) createBackupTasks(
	ctx context.Context,
	cfg *config.Config,
) {
	client := clients.NewFromAddress(s.GRPCAddr(), user.SystemTokenValue())
	ctx = user.SystemForGateway(ctx)
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
					user.SystemForLocal(context.Background()),
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
	ctx context.Context,
	db *taorm.DB,
	cfg *config.Config,
	sr grpc.ServiceRegistrar,
	cancel func(),
	filesStore *storage.SQLite,
	fileCache *cache.FileCache,
	rc *runtime_config.Runtime,
	mux *http.ServeMux,
) {
	serviceOptions := []service.With{
		service.WithPostDataFileSystem(filesStore),
		service.WithNotifier(s.notifyServer),
		service.WithCancel(cancel),
		service.WithFileCache(fileCache),
	}

	s.main = service.New(ctx, sr, cfg, db, rc, mux, s.Auth(), serviceOptions...)
}

func (s *Server) createUtilsService(ctx context.Context, cfg *config.Config, sr grpc.ServiceRegistrar) {
	options := []micros_utils.Option{
		micros_utils.WithTimezone(cfg.Site.GetTimezoneLocation),
	}

	if key := cfg.Others.Geo.GeoDe.Key; key != `` {
		options = append(options, micros_utils.WithGaoDe(key))
	}

	s.utils = micros_utils.New(ctx, sr, options...)
}

func (s *Server) createAuthServices(ctx context.Context, cfg *config.Config, sr grpc.ServiceRegistrar, db *taorm.DB) {
	s.clientLoginService = client_login.NewClientLoginService(ctx, sr, cfg.Site.GetHome)
	s.userManager = auth.NewUsersService(ctx, db, sr)
	s.authFrontend = auth.NewAuth(db, cfg.Site.GetHome, cfg.Site.GetName, s.userManager)
	s.authMiddleware.SetAuth(s.authFrontend)
	passkeys.NewPasskeysService(ctx, sr, s.authFrontend.GetWA, s.authFrontend.GetUserByID)
}

func (s *Server) createNotifyService(ctx context.Context, db *taorm.DB, cfg *config.Config, sr grpc.ServiceRegistrar) {
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
					n.SendInstant(user.SystemForLocal(ctx), &proto.SendInstantRequest{
						Title:       `有堆积的日志未处理`,
						Body:        fmt.Sprintf(`条数：%d`, c),
						Immediately: true,
					})
				}
			}
		}
	}()

	s.notifyServer = n
}

// 因为 GRPC 服务启动后不能注册，所以返回了一个函数用于适时启动。
func (s *Server) serveGRPC(ctx context.Context) (func(), grpc.ServiceRegistrar) {
	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(100<<20),
		grpc.MaxSendMsgSize(100<<20),
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(exceptionRecoveryHandler)),
			s.authMiddleware.UserFromGatewayUnaryInterceptor(),
			s.authMiddleware.UserFromClientTokenUnaryInterceptor(),
			func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				if s.throttlerEnabled.Load() && s.throttler != nil {
					return s.throttler(ctx, req, info, handler)
				}
				return handler(ctx, req)
			},
			grpcLoggerUnary,
		),
		grpc_middleware.WithStreamServerChain(
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(exceptionRecoveryHandler)),
			s.authMiddleware.UserFromGatewayStreamInterceptor(),
			s.authMiddleware.UserFromClientTokenStreamInterceptor(),
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

// 运行 HTTP 服务。
// 真实地址回写到 s.httpAddr 中。
func (s *Server) serveHTTP(ctx context.Context, addr string, h http.Handler) {
	debugRequest := func(w http.ResponseWriter, r *http.Request) {
		// if strings.HasPrefix(r.URL.Path, `/v3/posts/1801/files`) {
		// 	log.Println(1)
		// }
		h.ServeHTTP(w, r)
	}

	server := &http.Server{
		Addr: addr,
		Handler: utils.ChainFuncs(
			http.Handler(http.HandlerFunc(debugRequest)),
			// 注意这个拦截器的能力：
			//
			// 所有进入服务端认证信息均被包含在 context 中，
			// 这也包含了 Gateway。
			//
			// 但是，gateway 虽然有了 auth context，但是如果使用的是 grpc-client，
			// 无法传递给 server，会再次用 auth.NewContextForRequestAsGateway 再度解析并传递。
			s.authMiddleware.UserFromCookieHandler,
			logs.NewRequestLoggerHandler(`access.log`),
			s.main.MaintenanceMode().Handler(func(ctx context.Context, r *http.Request) bool {
				return user.Context(ctx).User.IsAdmin() || strings.HasPrefix(r.URL.Path, `/debug/`)
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
	if s.testing {
		s.Main().TestingSetHTTPAddr(s.JoinPath())
	}
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
	ac := user.Context(ctx)
	log.Println(method, ac.UserAgent)
}

func (s *Server) initSyncs(ctx context.Context, cfg *config.Config, filesStore theme_fs.FS) {
	for _, backend := range []struct {
		config  *config.OSSConfigWithEnabled
		name    string
		store   string
		country string
	}{
		{
			config:  &cfg.Site.Sync.Aliyun,
			name:    `aliyun`,
			store:   `site.sync.aliyun`,
			country: `china`,
		},
		{
			config:  &cfg.Site.Sync.COS,
			name:    `cos`,
			store:   `site.sync.cos`,
			country: `china`,
		},
		{
			config:  &cfg.Site.Sync.R2,
			name:    `r2`,
			store:   `site.sync.r2`,
			country: `america`,
		},
		{
			config: &cfg.Site.Sync.Minio,
			name:   `minio`,
			store:  `site.sync.minio`,
		},
	} {
		if backend.config.Enabled {
			oss, err := sync_files.NewSyncToOSS(
				ctx, backend.name, &backend.config.OSSConfig, s.Main(),
				s.Main().GetPluginStorage(backend.store),
				filesStore,
			)
			if err != nil {
				s.sendNotify(`启动同步失败`, err.Error())
				log.Println(err)
				continue
			}
			s.Main().RegisterFileURLGetter(backend.name, _OssWithCountry{oss, backend.country})
			log.Println(`启动同步：`, backend.name)
		}
	}
}

type _OssWithCountry struct {
	*sync_files.SyncToOSS
	country string
}

func (oss _OssWithCountry) GetCountry() string {
	return oss.country
}

func (s *Server) initSubTasks(ctx context.Context, cfg *config.Config, filesStore *storage.SQLite) {
	if s.initLiveCheck {
		go live_check.LiveCheck(ctx, s.Main(), s.sendNotify)
	}

	if s.initRssTasks {
		s.initRSS()
	}

	if s.initBackupTasks {
		go s.createBackupTasks(ctx, cfg)
	}

	if s.initGitSyncTask {
		go git_repo.Sync(ctx,
			clients.NewFromAddress(s.GRPCAddr(), user.SystemTokenValue()),
			s.notifyServer, s.main.SetLastSyncAt,
		)
	}

	if s.initMonitorDomain {
		s.Main().SetDomainDays(-1)
		go expiration.MonitorDomain(
			ctx, cfg.Site.GetHome, s.notifyServer, cfg.Others.Whois.ApiLayer.Key, s.initMonitorDomainDelay,
			s.Main().SetDomainDays,
		)
	}

	if s.initMonitorCerts {
		s.Main().SetCertDays(-1)
		go expiration.MonitorCert(
			ctx, cfg.Site.GetHome, s.notifyServer,
			s.Main().SetCertDays,
		)
	}

	if s.initYearProgress {
		year_progress.New(ctx, s.Main().CalenderService())
	}

	s.initSyncs(ctx, cfg, filesStore)

	createImageBorderTask := func() *auto_image_border.Task {
		return auto_image_border.NewTask(
			s.Main().GetPluginStorage(`auto_image_border`),
			filesStore, func(id int) {
				s.Main().InvalidatePost(id)
			})
	}

	s.utils.SetAutoImageBorderCreator(createImageBorderTask)
}
