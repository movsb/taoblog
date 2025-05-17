package migration

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/movsb/taorm"
)

func Migrate(gdb, files, cache *sql.DB) {
	if err := gdb.Ping(); err != nil {
		panic(err)
	}

	var sVer string
	gdb.QueryRow(`SELECT sqlite_version()`).Scan(&sVer)
	log.Println(`SQLite Version:`, sVer)

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
		log.Fatalln("unknown database version")
	}
	if begin == len(gVersions) {
		return
	}

	taorm.NewDB(gdb).MustTxCall(func(txPosts *taorm.DB) {
		taorm.NewDB(files).MustTxCall(func(txFiles *taorm.DB) {
			taorm.NewDB(cache).MustTxCall(func(txCache *taorm.DB) {
				for ; begin < len(gVersions); begin++ {
					v := gVersions[begin]
					if v.update != nil {
						log.Printf("updating to DB version %d ...\n", v.version)
						switch typed := v.update.(type) {
						case func(*sql.Tx):
							panic(`no longer supported`)
						case func(*taorm.DB, *taorm.DB, *taorm.DB):
							typed(txPosts, txFiles, txCache)
						default:
							panic(`unknown update function`)
						}
					}
				}
				lastVer := gVersions[len(gVersions)-1]
				txPosts.MustExec(`UPDATE options SET VALUE=? WHERE name='db_ver'`, lastVer.version)
			})
		})
	})
}

func mustExec(tx *sql.Tx, query string, args ...any) {
	_, err := tx.Exec(query, args...)
	if err != nil {
		panic(err)
	}
}
