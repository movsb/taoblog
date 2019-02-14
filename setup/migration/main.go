package migration

import (
	"database/sql"
	"fmt"
	"strconv"
)

// Migrate ...
func Migrate(gdb *sql.DB) {
	if err := gdb.Ping(); err != nil {
		panic(err)
	}
	row := gdb.QueryRow(`SELECT value FROM options WHERE name='db_ver'`)
	strDBVer := ""
	dbVer := 0
	if err := row.Scan(&strDBVer); err != nil {
		if err == sql.ErrNoRows {
			gdb.Exec(`INSERT INTO options (name,value) VALUES ('db_ver',?)`, 0)
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
			`UPDATE options SET VALUE=? WHERE name='db_ver'`,
			fmt.Sprint(lastVer.version),
		); err != nil {
			panic(err)
		}
	})
}
