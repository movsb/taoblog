package main

import (
	"database/sql"
)

func v0(tx *sql.Tx) {

}

func v1(tx *sql.Tx) {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()
	_, err = tx.Exec(`UPDATE posts SET source='' WHERE source IS NULL`)
	_, err = tx.Exec(`ALTER TABLE posts CHANGE source source LONGTEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL`)
}
