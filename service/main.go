package service

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/modules/taorm"
)

// ImplServer implements IServer.
type ImplServer struct {
	db   *sql.DB
	tdb  *taorm.DB
	auth *auth.Auth
}

// NewImplServer ...
func NewImplServer(db *sql.DB, auth *auth.Auth) *ImplServer {
	s := &ImplServer{
		db:   db,
		tdb:  taorm.NewDB(db),
		auth: auth,
	}
	return s
}

func joinInts(ints []int64, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ints)), delim), "[]")
}
