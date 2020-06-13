package service

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/memory_cache"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/file_managers"
	"github.com/movsb/taorm/taorm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	grpcAddress = "127.0.0.1:2563"
)

// IFileManager exposes interfaces to manage upload files.
type IFileManager interface {
	Put(pid int64, name string, r io.Reader) error
	Delete(pid int64, name string) error
	List(pid int64) ([]string, error)
}

// Service implements IServer.
type Service struct {
	cfg    *config.Config
	tdb    *taorm.DB
	auth   *auth.Auth
	cmtntf *comment_notify.CommentNotifier
	fmgr   IFileManager
	cache  *memory_cache.MemoryCache
}

// NewService ...
func NewService(cfg *config.Config, db *sql.DB, auth *auth.Auth) *Service {
	s := &Service{
		cfg:   cfg,
		tdb:   taorm.NewDB(db),
		auth:  auth,
		fmgr:  file_managers.NewLocalFileManager(cfg.Data.File.Path),
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
			grpc_recovery.UnaryServerInterceptor(
				grpc_recovery.WithRecoveryHandler(exceptionRecoveryHandler),
			),
		),
	)

	protocols.RegisterTaoBlogServer(server, s)

	listener, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		panic(err)
	}
	go server.Serve(listener)

	return s
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
	closed := s.GetDefaultStringOption("site_closed", "false")
	if closed == "1" || closed == "true" {
		return true
	}
	if s.cfg.Maintenance.SiteClosed {
		return true
	}
	return false
}

// HomeURL returns the home URL of format https://localhost.
func (s *Service) HomeURL() string {
	return s.cfg.Blog.Home
}

func (s *Service) Name() string {
	return s.cfg.Blog.Name
}

func (s *Service) Description() string {
	return s.cfg.Blog.Description
}

/*
// temporary
func (s *Service) Add304(w http.ResponseWriter, req *http.Request) {
	// TODO turn on 304 or off from config.
	if modified := s.GetDefaultStringOption("last_post_time", ""); modified != "" {
		w.Header().Add("Last-Modified", datetime.My2Gmt(modified))
	}
	s.addVersionCookie(w)
}

func (s *Service) Handle304(w http.ResponseWriter, req *http.Request, date time.Time) bool {
	var notModified bool
	var commitMatched bool

	// TODO simplify
	if modified := req.Header.Get(`If-Modified-Since`); modified != "" {
		if datetime.My2Gmt(date.Format(`2006-01-02 15:04:05`)) == modified {
			notModified = true
		}
	}

	if s.checkVersionCookie(req) {
		commitMatched = true
	}

	if notModified && commitMatched {
		w.WriteHeader(http.StatusNotModified)
		return true
	}

	return false
}

func (s *Service) addVersionCookie(w http.ResponseWriter) {
	v := version.GitCommit
	if v == `` {
		v = `HEAD`
	}
	http.SetCookie(w, &http.Cookie{
		Name:     `commit`,
		Value:    v,
		Path:     `/`,
		MaxAge:   0,
		Secure:   false,
		HttpOnly: true,
	})
}

func (s *Service) checkVersionCookie(req *http.Request) bool {
	commit, err := req.Cookie(`commit`)
	if err != nil {
		return false
	}
	return commit.Value == version.GitCommit
}

func (s *Service) LastUpdate() time.Time {
	modified := s.GetDefaultStringOption("last_post_time", "")
	t, err := time.Parse(`2006-01-02 15:04:05`, modified)
	if err != nil {
		return time.Now()
	}
	return t
}
*/
