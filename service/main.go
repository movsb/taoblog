package service

import (
	"database/sql"
	"io"
	"os"
	"strings"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/file_managers"
)

// IFileManager exposes interfaces to manage upload files.
type IFileManager interface {
	Put(pid int64, name string, r io.Reader) error
	Delete(pid int64, name string) error
	List(pid int64) ([]string, error)
}

// ImplServer implements IServer.
type ImplServer struct {
	tdb    *taorm.DB
	auth   *auth.Auth
	cmtntf *comment_notify.CommentNotifier
	fmgr   IFileManager
}

// NewImplServer ...
func NewImplServer(db *sql.DB, auth *auth.Auth) *ImplServer {
	s := &ImplServer{
		tdb:  taorm.NewDB(db),
		auth: auth,
		fmgr: file_managers.NewLocalFileManager(),
	}
	mailConfig := strings.SplitN(os.Getenv("MAIL"), "/", 3)
	if len(mailConfig) != 3 {
		panic("bad mail")
	}

	s.cmtntf = &comment_notify.CommentNotifier{
		MailServer: mailConfig[0],
		Username:   mailConfig[1],
		Password:   mailConfig[2],
		AdminName:  s.GetDefaultStringOption("author", ""),
		AdminEmail: s.GetDefaultStringOption("email", ""),
	}

	return s
}

// TxCall ...
func (s *ImplServer) TxCall(callback func(txs *ImplServer) error) {
	err := s.tdb.TxCall(func(tx *taorm.DB) error {
		txs := *s
		txs.tdb = tx
		return callback(&txs)
	})
	if err != nil {
		panic(err)
	}
}
