package taorm

import "database/sql"

func (db *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return db.db.Exec(query, args...)
}

func (db *DB) MustExec(query string, args ...interface{}) sql.Result {
	result, err := db.Exec(query, args...)
	if err != nil {
		panic(err)
	}
	return result
}
