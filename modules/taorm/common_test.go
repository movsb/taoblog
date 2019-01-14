package taorm

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

const mysqlTimeFormat = "2006-01-02 15:04:05"

type Post struct {
	ID       int64
	Date     string
	Modified string
	Title    string
	Content  string
	Status   int
	Metas    string
}

func dbConn() *sql.DB {
	dataSource := fmt.Sprintf("%[1]s:%[1]s@/%[1]s", "taoblog")
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}

	return db
}

func mustExec(db *sql.DB, query string, args ...interface{}) {
	if _, err := db.Exec(query, args...); err != nil {
		panic(err)
	}
}

func TestDropTables(t *testing.T) {
	db := dbConn()
	defer db.Close()
	mustExec(db, `drop table posts`)
	mustExec(db, `drop table comments`)
}

func TestCreateTables(t *testing.T) {
	db := dbConn()
	defer db.Close()
	{
		/*mustExec(db, `CREATE TABLE comments (
			id int(20) unsigned NOT NULL AUTO_INCREMENT,
			post_id int(20) unsigned NOT NULL,
			author tinytext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
			email varchar(100) NOT NULL,
			url varchar(200) DEFAULT NULL,
			ip varchar(16) NOT NULL,
			date datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
			content text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
			parent int(20) unsigned NOT NULL,
			ancestor int(20) unsigned NOT NULL,
			PRIMARY KEY (id)
		  )`)*/
	}
}

func TestImportPosts(t *testing.T) {
	//fp, err := os.Open("")
}
