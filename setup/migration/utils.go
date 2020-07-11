package migration

import "database/sql"

func txCall(db *sql.DB, callback func(tx *sql.Tx)) {
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	catchCall := func() (except interface{}) {
		defer func() {
			except = recover()
		}()
		callback(tx)
		return
	}
	if except := catchCall(); except != nil {
		tx.Rollback()
		panic(except)
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
	}
}

func mustExec(tx *sql.Tx, query string, args ...interface{}) {
	if _, err := tx.Exec(query, args...); err != nil {
		panic(err)
	}
}
