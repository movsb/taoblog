package service

import (
	"database/sql"
	"io"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/modules/taorm"
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
	tdb  *taorm.DB
	auth *auth.Auth
	fmgr IFileManager
}

// NewImplServer ...
func NewImplServer(db *sql.DB, auth *auth.Auth) *ImplServer {
	s := &ImplServer{
		tdb:  taorm.NewDB(db),
		auth: auth,
		fmgr: file_managers.NewLocalFileManager(),
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
