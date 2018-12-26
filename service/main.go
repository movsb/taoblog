package service

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/movsb/taoblog/modules/taorm"
)

// ImplServer implements IServer.
type ImplServer struct {
	db  *sql.DB
	tdb *taorm.DB
}

// NewImplServer ...
func NewImplServer(db *sql.DB) *ImplServer {
	s := &ImplServer{
		db: db,
	}
	return s
}

func joinInts(ints []int64, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ints)), delim), "[]")
}
