package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"
	"sync/atomic"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/memory_cache"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols"
	commentgeo "github.com/movsb/taoblog/service/modules/comment_geo"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/search"
	"github.com/movsb/taorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Service implements IServer.
type Service struct {
	cfg    *config.Config
	db     *sql.DB
	tdb    *taorm.DB
	auth   *auth.Auth
	cmtntf *comment_notify.CommentNotifier
	cmtgeo *commentgeo.CommentGeo
	cache  *memory_cache.MemoryCache

	avatarCache *AvatarCache

	// 搜索引擎启动需要时间，所以如果网站一运行即搜索，则可能出现引擎不可用
	// 的情况，此时此值为空。
	searcher atomic.Pointer[search.Engine]

	protocols.TaoBlogServer
	protocols.ManagementServer
	protocols.SearchServer
}

// NewService ...
func NewService(cfg *config.Config, db *sql.DB, auther *auth.Auth) *Service {
	s := &Service{
		cfg:   cfg,
		db:    db,
		tdb:   taorm.NewDB(db),
		auth:  auther,
		cache: memory_cache.NewMemoryCache(time.Minute * 10),
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
			s.auth.UserFromGatewayInterceptor(),
			s.auth.UserFromClientTokenUnaryInterceptor(),
			grpcLoggerUnary,
		),
		grpc_middleware.WithStreamServerChain(
			grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(exceptionRecoveryHandler)),
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

func (s *Service) DevMode() bool {
	return version.GitCommit == "" || strings.EqualFold(version.GitCommit, `head`)
}

// LastArticleUpdateTime ...
func (s *Service) LastArticleUpdateTime() time.Time {
	t := time.Now()
	if modified := s.GetDefaultIntegerOption("last_post_time", 0); modified > 0 {
		t = time.Unix(modified, 0)
	}
	return t
}
