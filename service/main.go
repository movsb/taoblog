package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/movsb/pkg/notify"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/modules/cache"
	commentgeo "github.com/movsb/taoblog/service/modules/comment_geo"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/search"
	"github.com/movsb/taorm"
	"github.com/phuslu/lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// 请求节流器限流信息。
// 由于没有用户系统，目前根据 IP 限流。
// 这样会对网吧、办公网络非常不友好。
type _RequestThrottlerKey struct {
	UserID int
	IP     netip.Addr
	Method string // 指 RPC 方法，用路径代替。
}

func throttlerKeyOf(ctx context.Context) _RequestThrottlerKey {
	ac := auth.Context(ctx)
	method, ok := grpc.Method(ctx)
	if !ok {
		panic(status.Error(codes.Internal, "没有找到调用方法。"))
	}
	return _RequestThrottlerKey{
		UserID: int(ac.User.ID),
		IP:     ac.RemoteAddr,
		Method: method,
	}
}

// Service implements IServer.
type Service struct {
	cfg    *config.Config
	db     *sql.DB
	tdb    *taorm.DB
	auth   *auth.Auth
	cmtntf *comment_notify.CommentNotifier
	cmtgeo *commentgeo.CommentGeo

	// 通用缓存
	cache *lru.TTLCache[string, any]

	// 文章内容缓存。
	// NOTE：缓存 Key 是跟文章编号和内容选项相关的（因为内容选项不同内容也就不同），
	// 所以存在一篇文章有多个缓存的问题。所以后面单独加了一个缓存来记录一篇文章对应了哪些
	// 与之相关的内容缓存。创建/删除内容缓存时，同步更新两个缓存。
	// TODO 我是不是应该直接缓存 *Post？不过好像也挺好改的。
	postContentCaches *lru.TTLCache[_PostContentCacheKey, string]
	postCaches        *cache.RelativeCacheKeys[int64, _PostContentCacheKey]

	// 请求节流器。
	throttler *utils.Throttler[_RequestThrottlerKey]

	avatarCache *AvatarCache

	// 搜索引擎启动需要时间，所以如果网站一运行即搜索，则可能出现引擎不可用
	// 的情况，此时此值为空。
	searcher atomic.Pointer[search.Engine]

	maintenance *utils.Maintenance

	protocols.TaoBlogServer
	protocols.ManagementServer
	protocols.SearchServer
}

// NewService ...
func NewService(cfg *config.Config, db *sql.DB, auther *auth.Auth) *Service {
	s := &Service{
		cfg:  cfg,
		db:   db,
		tdb:  taorm.NewDB(db),
		auth: auther,

		cache:             lru.NewTTLCache[string, any](1024),
		postContentCaches: lru.NewTTLCache[_PostContentCacheKey, string](1024),
		postCaches:        cache.NewRelativeCacheKeys[int64, _PostContentCacheKey](),
		throttler:         utils.NewThrottler[_RequestThrottlerKey](),
		maintenance:       &utils.Maintenance{},
	}

	s.cmtntf = &comment_notify.CommentNotifier{
		MailServer: s.cfg.Server.Mailer.Server,
		Username:   s.cfg.Server.Mailer.Account,
		Password:   s.cfg.Server.Mailer.Password,
		Config:     &s.cfg.Comment,
	}
	s.cmtntf.Init()

	s.avatarCache = NewAvatarCache()
	s.cmtgeo = commentgeo.NewCommentGeo(context.TODO())

	s.cacheAllCommenterData()

	server := grpc.NewServer(
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

	protocols.RegisterTaoBlogServer(server, s)
	protocols.RegisterManagementServer(server, s)
	protocols.RegisterSearchServer(server, s)

	listener, err := net.Listen("tcp", cfg.Server.GRPCListen)
	if err != nil {
		panic(err)
	}
	go server.Serve(listener)

	go s.RunSearchEngine(context.TODO())
	go s.monitorCert(notify.NewOfficialChanify(s.cfg.Comment.Push.Chanify.Token))

	return s
}

// 从 Context 中取出用户并且必须为 Admin，否则 panic。
func (s *Service) MustBeAdmin(ctx context.Context) *auth.AuthContext {
	ac := auth.Context(ctx)
	if ac == nil {
		panic("AuthContext 不应为 nil")
	}
	if !ac.User.IsAdmin() {
		panic(status.Error(codes.PermissionDenied, "此操作无权限。"))
	}
	return ac
}

// GrpcAddress ...
func (s *Service) GrpcAddress() string {
	return s.cfg.Server.GRPCListen
}

func grpcLoggerUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	grpcLogger(ctx, info.FullMethod)
	return handler(ctx, req)
}
func grpcLoggerStream(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	grpcLogger(ss.Context(), info.FullMethod)
	return handler(srv, ss)
}
func grpcLogger(ctx context.Context, method string) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Println(md)
	}
	ac := auth.Context(ctx)
	log.Println(method, ac.UserAgent)
}

func exceptionRecoveryHandler(e interface{}) error {
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
	return status.New(codes.Internal, fmt.Sprint(e)).Err()
}

// Ping ...
func (s *Service) Ping(ctx context.Context, in *protocols.PingRequest) (*protocols.PingResponse, error) {
	return &protocols.PingResponse{
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

// HomeURL returns the home URL of format https://localhost.
func (s *Service) HomeURL() string {
	return strings.TrimSuffix(s.cfg.Site.Home, "/")
}

func (s *Service) Name() string {
	return s.cfg.Site.Name
}

func (s *Service) Description() string {
	if b := s.cfg.Site.ShowDescription; !b {
		return ""
	}
	if d := s.cfg.Site.Description; d != `` {
		return d
	}
	return ``
}

func DevMode() bool {
	return os.Getenv(`DEV`) != `0` && (version.GitCommit == "" || strings.EqualFold(version.GitCommit, `head`))
}

// 是否在维护模式。
// 1. 手动进入。
// 2. 自动升级过程中。
// https://github.com/movsb/taoblog/commit/c4428d7
func (s *Service) MaintenanceMode() *utils.Maintenance {
	return s.maintenance
}

// LastArticleUpdateTime ...
func (s *Service) LastArticleUpdateTime() time.Time {
	t := time.Now()
	if modified := s.GetDefaultIntegerOption("last_post_time", 0); modified > 0 {
		t = time.Unix(modified, 0)
	}
	return t
}

func (s *Service) CreateCustomThemeApplyFunc() func() string {
	c := s.cfg.Site.Theme.Stylesheets
	stylesheetTmpl := template.Must(template.New(`stylesheet`).Parse(c.Template))
	return func() string {
		w := bytes.NewBuffer(nil)
		for _, ss := range c.Stylesheets {
			stylesheetTmpl.Execute(w, ss)
			fmt.Fprintln(w)
		}
		return w.String()
	}
}

var methodThrottlerInfo = map[string]struct {
	Interval time.Duration
	Message  string

	// 仅节流返回正确错误码的接口。
	// 如果接口返回错误，不更新。
	OnSuccess bool
}{
	`/protocols.TaoBlog/CreateComment`: {
		Interval:  time.Second * 10,
		Message:   `评论发表过于频繁，请稍等几秒后再试。`,
		OnSuccess: true,
	},
	`/protocols.TaoBlog/UpdateComment`: {
		Interval:  time.Second * 5,
		Message:   `评论更新过于频繁，请稍等几秒后再试。`,
		OnSuccess: true,
	},
}

func (s *Service) throttlerGatewayInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	key := throttlerKeyOf(ctx)
	ti, ok := methodThrottlerInfo[info.FullMethod]
	if ok {
		if s.throttler.Throttled(key, ti.Interval, false) {
			msg := utils.IIF(ti.Message != "", ti.Message, `你被节流了，请稍候再试。You've been throttled.`)
			return nil, status.Error(codes.Aborted, msg)
		}
	}
	resp, err := handler(ctx, req)

	if !ti.OnSuccess || err == nil {
		s.throttler.Update(key, ti.Interval)
	}

	return resp, err
}

// 监控证书过期的剩余时间。
func (s *Service) monitorCert(chanify *notify.Chanify) {
	home := s.cfg.Site.Home
	u, err := url.Parse(home)
	if err != nil {
		panic(err)
	}
	if u.Scheme != `https` {
		return
	}
	port := utils.IIF(u.Port() == "", "443", u.Port())
	addr := net.JoinHostPort(u.Hostname(), port)
	check := func() {
		conn, err := tls.Dial(`tcp`, addr, &tls.Config{})
		if err != nil {
			log.Println(err)
			if chanify != nil {
				chanify.Send(`错误`, err.Error(), true)
			}
			return
		}
		defer conn.Close()
		cert := conn.ConnectionState().PeerCertificates[0]
		left := time.Until(cert.NotAfter)
		if left <= 0 {
			log.Println(`已过期`)
			if chanify != nil {
				chanify.Send(`证书`, `已经过期。`, true)
			}
			return
		}
		daysLeft := int(left.Hours() / 24)
		if daysLeft >= 15 {
			return
		}
		log.Println(`剩余天数：`, daysLeft)
		if chanify != nil {
			chanify.Send(`证书`, fmt.Sprintf(`剩余天数：%v`, daysLeft), true)
		}
	}
	check()
	go func() {
		ticker := time.NewTicker(time.Hour * 24)
		defer ticker.Stop()
		for range ticker.C {
			check()
		}
	}()
}
