package service

import (
	"database/sql"
	"fmt"
	"strings"
)

// ImplServer implements IServer.
type ImplServer struct {
	db   *sql.DB
	auth IAuth
}

// NewImplServer ...
func NewImplServer(db *sql.DB, auth IAuth) *ImplServer {
	s := &ImplServer{
		db:   db,
		auth: auth,
	}
	return s
}

func joinInts(ints []int64, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ints)), delim), "[]")
}
