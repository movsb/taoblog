package service

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/memory_cache"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/search"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taoblog/service/modules/storage/local"
	"github.com/movsb/taorm/taorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Service implements IServer.
type Service struct {
	cfg      *config.Config
	db       *sql.DB
	tdb      *taorm.DB
	auth     *auth.Auth
	cmtntf   *comment_notify.CommentNotifier
	store    storage.Store
	cache    *memory_cache.MemoryCache
	searcher *search.Engine
}

// NewService ...
func NewService(cfg *config.Config, db *sql.DB, auther *auth.Auth) *Service {
	localStorage, err := local.NewLocal(cfg.Data.File.Path)
	if err != nil {
		panic(err)
	}

	s := &Service{
		cfg:   cfg,
		db:    db,
		tdb:   taorm.NewDB(db),
		auth:  auther,
		store: localStorage,
		cache: memory_cache.NewMemoryCache(time.Minute * 10),
	}

	s.cmtntf = &comment_notify.CommentNotifier{
		MailServer: s.cfg.Server.Mailer.Server,
		Username:   s.cfg.Server.Mailer.Account,
		Password:   s.cfg.Server.Mailer.Password,
		AdminName:  s.cfg.Comment.Author,
		AdminEmail: s.cfg.Comment.Email,
		Config:     s.cfg,
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

// Store ...
func (s *Service) Store() storage.Store {
	return s.store
}

// TxCall ...
func (s *Service) TxCall(callback func(txs *Service) error) {
	err := s.tdb.TxCall(func(tx *taorm.DB) error {
		txs := *s
		txs.tdb = tx
		return callback(&txs)
	})
	if err != nil {
		panic(err)
	}
}

func (s *Service) IsSiteClosed() bool {
	return s.cfg.Maintenance.SiteClosed
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
	if n := len(s.cfg.Site.Mottoes); n > 0 {
		return s.cfg.Site.Mottoes[rand.Intn(n)]
	}
	return ``
}

// LastArticleUpdateTime ...
func (s *Service) LastArticleUpdateTime() time.Time {
	modified := s.GetDefaultStringOption("last_post_time", "")
	t, err := time.ParseInLocation(`2006-01-02 15:04:05`, modified, time.Local)
	if err != nil {
		return time.Now()
	}
	return t
}
