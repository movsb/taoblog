package service

import "database/sql"

// ImplServer implements IServer.
type ImplServer struct {
	db *sql.DB
}

// NewImplServer ...
func NewImplServer(db *sql.DB) *ImplServer {
	s := &ImplServer{
		db: db,
	}
	return s
}
