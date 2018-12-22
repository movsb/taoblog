package service

import "database/sql"

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
