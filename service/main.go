package service

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/memory_cache"
	"github.com/movsb/taoblog/protocols"
	commentgeo "github.com/movsb/taoblog/service/modules/comment_geo"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/search"
	"github.com/movsb/taoblog/theme/modules/canonical"
	"github.com/movsb/taorm/taorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

	linker canonical.Linker

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
		cfg:    cfg,
		db:     db,
		tdb:    taorm.NewDB(db),
		auth:   auther,
		cache:  memory_cache.NewMemoryCache(time.Minute * 10),
		cmtgeo: commentgeo.NewCommentGeo(context.TODO()),

		avatarCache: NewAvatarCache(),
	}

	s.cmtntf = &comment_notify.CommentNotifier{
		MailServer: s.cfg.Server.Mailer.Server,
		Username:   s.cfg.Server.Mailer.Account,
		Password:   s.cfg.Server.Mailer.Password,
		AdminName:  s.cfg.Comment.Author,
		AdminEmail: s.cfg.Comment.Emails[0], // TODO 如果没配置，则不启用此功能
		Config:     &s.cfg.Comment,
	}
	s.cmtntf.Init()

	server := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(exceptionRecoveryHandler)),
			auth.GatewayAuthInterceptor(s.auth),
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

func (s *Service) SetLinker(linker canonical.Linker) {
	s.linker = linker
}

// GrpcAddress ...
func (s *Service) GrpcAddress() string {
	return s.cfg.Server.GRPCListen
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
	return s.cfg.Site.Home
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

// LastArticleUpdateTime ...
func (s *Service) LastArticleUpdateTime() time.Time {
	t := time.Now()
	if modified := s.GetDefaultIntegerOption("last_post_time", 0); modified > 0 {
		t = time.Unix(modified, 0)
	}
	return t
}
