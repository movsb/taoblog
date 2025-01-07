package service

import (
	"context"
	"database/sql"
	"expvar"
	"fmt"
	"log"
	"net"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/mailer"
	"github.com/movsb/taoblog/modules/notify"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/cache"
	commentgeo "github.com/movsb/taoblog/service/modules/comment_geo"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/renderers/exif"
	"github.com/movsb/taoblog/service/modules/renderers/friends"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
	"github.com/movsb/taoblog/service/modules/search"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/movsb/taorm"
	"github.com/phuslu/lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ToBeImplementedByRpc interface {
	ListAllPostsIds(ctx context.Context) ([]int32, error)
	GetDefaultIntegerOption(name string, def int64) int64
	GetLink(ID int64) string
	GetPlainLink(ID int64) string
	Config() *config.Config
	ListTagsWithCount() []*models.TagWithCount
	IncrementPostPageView(id int64)
	ThemeChangedAt() time.Time
	GetCommentEmailById(id int) string
}

// Service implements IServer.
type Service struct {
	ctx    context.Context
	cancel context.CancelFunc

	testing bool
	addr    net.Addr // 服务器的实际监听地址（GRPC）

	// 计划重启。
	scheduledUpdate atomic.Bool

	home *url.URL

	// 服务器默认的时区。
	timeLocation *time.Location

	cfg *config.Config

	postDataFS theme_fs.FS

	db   *sql.DB
	tdb  *taorm.DB
	auth *auth.Auth

	mailer   *mailer.MailerLogger
	notifier notify.Notifier
	cmtntf   *comment_notify.CommentNotifier
	cmtgeo   *commentgeo.Task

	// 通用缓存
	cache *lru.TTLCache[string, any]
	// 图片元数据缓存任务。
	exifTask *exif.Task
	// 朋友头像数据缓存任务
	friendsTask *friends.Task
	// 提醒事项任务
	remindersTask *reminders.Task

	// 文章内容缓存。
	// NOTE：缓存 Key 是跟文章编号和内容选项相关的（因为内容选项不同内容也就不同），
	// 所以存在一篇文章有多个缓存的问题。所以后面单独加了一个缓存来记录一篇文章对应了哪些
	// 与之相关的内容缓存。创建/删除内容缓存时，同步更新两个缓存。
	// TODO 我是不是应该直接缓存 *Post？不过好像也挺好改的。
	postContentCaches    *lru.TTLCache[_PostContentCacheKey, string]
	postCaches           *cache.RelativeCacheKeys[int64, _PostContentCacheKey]
	commentContentCaches *lru.TTLCache[_PostContentCacheKey, string]
	commentCaches        *cache.RelativeCacheKeys[int64, _PostContentCacheKey]

	// 基本临时文件的缓存。
	filesCache *cache.TmpFiles

	// 请求节流器。
	throttler        grpc.UnaryServerInterceptor
	throttlerEnabled atomic.Bool

	avatarCache *cache.AvatarHash

	// 搜索引擎启动需要时间，所以如果网站一运行即搜索，则可能出现引擎不可用
	// 的情况，此时此值为空。
	searcher         atomic.Pointer[search.Engine]
	onceInitSearcher sync.Once

	maintenance *utils.Maintenance

	// 服务内有插件的更新可能会影响到内容渲染。
	themeChangedAt time.Time

	exporter *_Exporter
	// 证书剩余天数。
	// >= 0 表示值有效。
	certDaysLeft atomic.Int32
	// 域名有效期
	// >= 0 表示值有效。
	domainExpirationDaysLeft atomic.Int32

	proto.AuthServer
	proto.TaoBlogServer
	proto.ManagementServer
	proto.SearchServer
	proto.UtilsServer
}

func (s *Service) ThemeChangedAt() time.Time {
	return s.themeChangedAt
}

func (s *Service) Addr() net.Addr {
	return s.addr
}

func NewService(ctx context.Context, cancel context.CancelFunc, testing bool, cfg *config.Config, db *sql.DB, auther *auth.Auth, options ...With) *Service {
	return newService(ctx, cancel, cfg, db, auther, testing, options...)
}

func newService(ctx context.Context, cancel context.CancelFunc, cfg *config.Config, db *sql.DB, auther *auth.Auth, testing bool, options ...With) *Service {
	s := &Service{
		ctx:     ctx,
		cancel:  cancel,
		testing: testing,

		notifier: notify.NewConsoleNotify(),

		cfg:        cfg,
		postDataFS: &theme_fs.Empty{},

		// TODO 可配置使用的时区，而不是使用服务器当前时间或者硬编码成+8时区。
		timeLocation: time.Now().Location(),

		db:   db,
		tdb:  taorm.NewDB(db),
		auth: auther,

		cache:                lru.NewTTLCache[string, any](128),
		postContentCaches:    lru.NewTTLCache[_PostContentCacheKey, string](10240),
		postCaches:           cache.NewRelativeCacheKeys[int64, _PostContentCacheKey](),
		commentContentCaches: lru.NewTTLCache[_PostContentCacheKey, string](10240),
		commentCaches:        cache.NewRelativeCacheKeys[int64, _PostContentCacheKey](),
		filesCache:           cache.NewTmpFiles(".cache", time.Hour*24*7),

		maintenance: utils.NewMaintenance(),

		themeChangedAt: time.Now(),
	}

	for _, opt := range options {
		opt(s)
	}

	utilsService := NewUtils(s.notifier)
	s.UtilsServer = utilsService

	if u, err := url.Parse(cfg.Site.Home); err != nil {
		panic(err)
	} else {
		s.home = u
	}

	s.avatarCache = cache.NewAvatarHash()
	s.cmtntf = comment_notify.New(&s.cfg.Comment, s.notifier, s.mailer)
	s.cmtgeo = commentgeo.NewTask(s.GetPluginStorage(`cmt_geo`))

	s.cacheAllCommenterData()

	exifDebouncer := utils.NewBatchDebouncer(time.Second*10, func(id int) {
		s.deletePostContentCacheFor(int64(id))
		s.updatePostMetadataTime(int64(id), time.Now())
	})
	s.exifTask = exif.NewTask(s.GetPluginStorage(`exif`), func(id int) {
		exifDebouncer.Enter(id)
	})

	s.friendsTask = friends.NewTask(s.GetPluginStorage(`friends`), func(postID int) {
		s.deletePostContentCacheFor(int64(postID))
		s.updatePostMetadataTime(int64(postID), time.Now())
	})

	s.remindersTask = reminders.NewTask(s.GetPluginStorage(`reminders`), func(pid int64) {
		s.deletePostContentCacheFor(pid)
		s.updatePostMetadataTime(pid, time.Now())
	})

	s.certDaysLeft.Store(-1)
	s.domainExpirationDaysLeft.Store(-1)
	s.exporter = _NewExporter(s)

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

	proto.RegisterUtilsServer(server, s)
	proto.RegisterAuthServer(server, s)
	proto.RegisterTaoBlogServer(server, s)
	proto.RegisterManagementServer(server, s)
	proto.RegisterSearchServer(server, s)

	listener, err := net.Listen("tcp", `127.0.0.1:0`)
	if err != nil {
		panic(err)
	}
	s.addr = listener.Addr()
	log.Println(`GRPC listen on:`, listener.Addr().String())
	go server.Serve(listener)

	if !testing && !version.DevMode() {
		go s.monitorCert(s.notifier)
		go s.monitorDomain(s.notifier)
	}

	return s
}

// 从 Context 中取出用户并且必须为 Admin/System，否则 panic。
func (s *Service) MustBeAdmin(ctx context.Context) *auth.AuthContext {
	return MustBeAdmin(ctx)
}

func MustBeAdmin(ctx context.Context) *auth.AuthContext {
	ac := auth.Context(ctx)
	if ac == nil {
		panic("AuthContext 不应为 nil")
	}
	if !ac.User.IsAdmin() && !ac.User.IsSystem() {
		panic(status.Error(codes.PermissionDenied, "此操作无权限。"))
	}
	return ac
}

func (s *Service) throttlerGatewayInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if s.throttlerEnabled.Load() && s.throttler != nil {
		return s.throttler(ctx, req, info, handler)
	}
	return handler(ctx, req)
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

// Ping ...
func (s *Service) Ping(ctx context.Context, in *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{
		Pong: `pong`,
	}, nil
}

// Config ...
func (s *Service) Config() *config.Config {
	return s.cfg
}

// MustTxCall ...
func (s *Service) MustTxCall(callback func(txs *Service) error) {
	if err := s.TxCall(callback); err != nil {
		panic(err)
	}
}

// TxCall ...
func (s *Service) TxCall(callback func(txs *Service) error) error {
	return s.tdb.TxCall(func(tx *taorm.DB) error {
		txs := *s
		txs.tdb = tx
		return callback(&txs)
	})
}

// 是否在维护模式。
// 1. 手动进入。
// 2. 自动升级过程中。
// https://github.com/movsb/taoblog/commit/c4428d7
func (s *Service) MaintenanceMode() *utils.Maintenance {
	return s.maintenance
}

func (s *Service) GetInfo(ctx context.Context, in *proto.GetInfoRequest) (*proto.GetInfoResponse, error) {
	t := time.Now()
	if modified := s.GetDefaultIntegerOption("last_post_time", 0); modified > 0 {
		t = time.Unix(modified, 0)
	}

	out := &proto.GetInfoResponse{
		Name:         s.cfg.Site.Name,
		Description:  s.cfg.Site.Description,
		Home:         strings.TrimSuffix(s.cfg.Site.Home, "/"),
		LastPostedAt: int32(t.Unix()),

		CertDaysLeft:   s.certDaysLeft.Load(),
		DomainDaysLeft: s.domainExpirationDaysLeft.Load(),

		ScheduledUpdate: s.scheduledUpdate.Load(),
	}

	return out, nil
}
