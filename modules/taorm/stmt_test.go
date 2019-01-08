package taorm

import (
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	db := dbConn()
	defer db.Close()
	mustExec(db, `CREATE TABLE IF NOT EXISTS posts (
		id int(20) unsigned NOT NULL AUTO_INCREMENT,
		date datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
		modified datetime NOT NULL DEFAULT '1970-01-01 00:00:00',
		title text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
		content longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
		status int(10) NOT NULL,
		metas text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
		PRIMARY KEY (id)
		)`,
	)

	tdb := NewDB(db)

	var p Post
	p.ID = 100
	p.Date = time.Now().Format(mysqlTimeFormat)
	p.Modified = p.Date
	tdb.Model(&p, "posts").MustCreate()

	mustExec(db, `DROP TABLE posts`)
}
