package taorm

import "database/sql"

type _SQLCommon interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Expr is raw SQL string.
type Expr string
