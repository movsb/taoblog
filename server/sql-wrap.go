package main

import (
	"database/sql"
)

// From: https://github.com/jmoiron/sqlx/issues/344#issuecomment-318372779

// Querier is implemented by sql.DB and sql.Tx
type Querier interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func txCall(db *sql.DB, callback func(tx Querier) error) error {
	var err error

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if err = callback(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
