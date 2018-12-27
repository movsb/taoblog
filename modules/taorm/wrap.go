package taorm

import "database/sql"

// Querier is implemented by sql.DB and sql.Tx
// From: https://github.com/jmoiron/sqlx/issues/344#issuecomment-318372779
type Querier interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
