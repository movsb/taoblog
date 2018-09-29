package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

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

func main() {
	username := flag.String("username", "", "mysql username")
	password := flag.String("password", "", "mysql password")
	database := flag.String("database", "", "mysql database")
	flag.Parse()
	dataSource := fmt.Sprintf("%s:%s@/%s", *username, *password, *database)
	gdb, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	if err := gdb.Ping(); err != nil {
		panic(err)
	}
	defer gdb.Close()
	row := gdb.QueryRow(`SELECT value FROM options WHERE name='db_ver'`)
	strDBVer := ""
	dbVer := 0
	if err := row.Scan(&strDBVer); err != nil {
		if err == sql.ErrNoRows {
			dbVer = 0
		} else {
			panic(err)
		}
	} else {
		dbVer, err = strconv.Atoi(strDBVer)
		if err != nil {
			panic(err)
		}
	}
	begin := -1
	for i, v := range gVersions {
		if v.version == dbVer {
			begin = i + 1
			break
		}
	}
	if begin == -1 {
		panic("unknown version")
	}
	if begin == len(gVersions) {
		return
	}

	txCall(gdb, func(tx *sql.Tx) {
		for ; begin < len(gVersions); begin++ {
			v := gVersions[begin]
			if v.updater != nil {
				fmt.Printf("updating to DB version %d ...\n", v.version)
				v.updater(tx)
			}
		}
		lastVer := gVersions[len(gVersions)-1]
		if _, err := tx.Exec(
			`UPDATE options SET VALUE=? WHERE name='version'`,
			fmt.Sprint(lastVer.version),
		); err != nil {
			panic(err)
		}
	})
}
