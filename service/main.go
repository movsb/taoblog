package service

import (
	"database/sql"
	"io"
	"time"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/memory_cache"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/file_managers"
	"github.com/movsb/taorm/taorm"
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
		AdminName:  s.GetDefaultStringOption("author", ""),
		AdminEmail: s.GetDefaultStringOption("email", ""),
	}

	return s
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
	const homeURLKey = `home_url`
	if val, ok := s.cache.Get(homeURLKey); ok {
		return val.(string)
	}
	// TODO warn the owner if home isn't set
	home := "https://" + s.GetDefaultStringOption("home", "localhost")
	s.cache.Set(homeURLKey, home)
	return home
}
