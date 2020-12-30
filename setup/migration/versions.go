package migration

import (
	"database/sql"
	"encoding/json"

	"github.com/movsb/taorm/taorm"
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

func v2(tx *sql.Tx) {
	if _, err := tx.Exec(`DROP TABLE IF EXISTS shuoshuo;`); err != nil {
		panic(err)
	}
	if _, err := tx.Exec(`DROP TABLE IF EXISTS shuoshuo_comments;`); err != nil {
		panic(err)
	}
}

func v3(tx *sql.Tx) {
	type v3PostMetas struct {
		ID       int64
		Type     string
		Tid      int64
		Header   string
		Footer   string
		Keywords string
	}
	type v3Meta struct {
		ID    int64  `json:"id"`
		Metas string `json:"metas"`
	}
	var metas []*v3PostMetas
	if err := taorm.ScanRows(&metas, tx, `SELECT * FROM post_metas WHERE type='post' OR type='page'`); err != nil {
		panic(err)
	}
	for _, meta := range metas {
		var pm v3Meta
		query := `SELECT id,metas FROM posts WHERE id=?`
		if err := taorm.ScanRows(&pm, tx, query, meta.Tid); err != nil {
			panic(err)
		}
		var m map[string]interface{}
		if pm.Metas == "" {
			pm.Metas = "{}"
		}
		if err := json.Unmarshal([]byte(pm.Metas), &m); err != nil {
			panic(err)
		}
		m["header"] = meta.Header
		m["footer"] = meta.Footer
		by, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}
		pm.Metas = string(by)
		_, err = tx.Exec(`UPDATE posts SET metas=? WHERE id=?`, pm.Metas, pm.ID)
		if err != nil {
			panic(err)
		}
	}
	_, err := tx.Exec(`DROP TABLE post_metas`)
	if err != nil {
		panic(err)
	}
	_, err = tx.Exec(`UPDATE posts SET metas='{}' WHERE metas=''`)
	if err != nil {
		panic(err)
	}
}

func v4(tx *sql.Tx) {
	s := "CREATE UNIQUE INDEX `uix_post_id_and_tag_id` ON `post_tags` (`post_id`, `tag_id`)"
	if _, err := tx.Exec(s); err != nil {
		panic(err)
	}
}

func v5(tx *sql.Tx) {
	s := "CREATE UNIQUE INDEX `uix_name` ON `options` (`name`)"
	if _, err := tx.Exec(s); err != nil {
		panic(err)
	}
}

func v6(tx *sql.Tx) {
	ss := []string{
		"UPDATE posts SET date=DATE_ADD(date, INTERVAL 8 HOUR),modified=DATE_ADD(modified, INTERVAL 8 HOUR)",
		"UPDATE comments SET date=DATE_ADD(date, INTERVAL 8 HOUR)",
	}
	for _, s := range ss {
		if _, err := tx.Exec(s); err != nil {
			panic(err)
		}
	}
}

func v7(tx *sql.Tx) {
	var login string
	query := "SELECT value FROM options WHERE name=?"
	row := tx.QueryRow(query, "login")
	if err := row.Scan(&login); err != nil {
		panic(err)
	}
	query = "UPDATE options SET value=? WHERE name=?"
	if _, err := tx.Exec(query, login, "login"); err != nil {
		panic(err)
	}
}

func v8(tx *sql.Tx) {
	q := `ALTER TABLE comments CHANGE ancestor root INT(20) UNSIGNED NOT NULL`
	if _, err := tx.Exec(q); err != nil {
		panic(err)
	}
}

func v9(tx *sql.Tx) {
	q := `ALTER TABLE taxonomies CHANGE ancestor root INT(20) UNSIGNED NOT NULL`
	if _, err := tx.Exec(q); err != nil {
		panic(err)
	}
}

func v10(tx *sql.Tx) {
	queries := []string{
		"RENAME TABLE `taxonomies` TO `categories`",
		"ALTER TABLE `categories` ADD COLUMN `path` VARCHAR(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL",
	}
	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			panic(err)
		}
	}

	type v10Category struct {
		ID     uint
		Name   string
		Slug   string
		Parent uint
		Root   uint
		Path   string
	}

	var setPaths func(cat *v10Category, path string)
	setPaths = func(cat *v10Category, path string) {
		query := `UPDATE categories SET path=? WHERE id=?`
		if _, err := tx.Exec(query, path, cat.ID); err != nil {
			panic(err)
		}
		var children []*v10Category
		query = `SELECT * FROM categories WHERE parent=?`
		taorm.MustScanRows(&children, tx, query, cat.ID)
		childPath := path + "/" + cat.Slug
		if path == "/" {
			childPath = childPath[1:]
		}
		for _, child := range children {
			setPaths(child, childPath)
		}
	}

	var topLevels []*v10Category
	q := `SELECT * FROM categories WHERE parent = 0`
	taorm.MustScanRows(&topLevels, tx, q)
	for _, cat := range topLevels {
		setPaths(cat, "/")
	}

	queries = []string{
		"ALTER TABLE `categories` CHANGE `parent` `parent_id` INT(10) UNSIGNED NOT NULL",
		"ALTER TABLE `categories` DROP `root`",
		"DELETE FROM `categories` WHERE `path` = ''",
		"CREATE INDEX `uix_path_slug` ON `categories` (`path`,`slug`)",
		"ALTER TABLE `posts` CHANGE `taxonomy` `category` INT(10) UNSIGNED NOT NULL DEFAULT 1",
	}
	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			panic(err)
		}
	}
}

func v11(tx *sql.Tx) {
	tx.Exec(`DELETE FROM options WHERE name = ?`, `home`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `blog_name`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `blog_desc`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `keywords`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `login`)
}

func v12(tx *sql.Tx) {
	tx.Exec("ALTER TABLE comments ADD COLUMN `source_type` varchar(16) NOT NULL AFTER `date`")
	tx.Exec("ALTER TABLE comments ADD COLUMN `source` TEXT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL AFTER `source_type`")
	tx.Exec("UPDATE comments SET source_type='plain'")
}

func v13(tx *sql.Tx) {
	tx.Exec(`DELETE FROM options WHERE name = ?`, `email`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `not_allowed_emails`)
	tx.Exec(`DELETE FROM options WHERE name = ?`, `not_allowed_authors`)
}

func v14(tx *sql.Tx) {
	tx.Exec(`UPDATE posts SET comments = 0`)
	tx.Exec(`UPDATE posts INNER JOIN (SELECT post_id,count(id) AS comments FROM comments GROUP BY post_id) AS counts ON posts.id = counts.post_id SET posts.comments = counts.comments`)
}

func v15(tx *sql.Tx) {
	if _, err := tx.Exec(`DELETE FROM categories WHERE slug=? AND parent_id=?`, `uncategorized`, 0); err != nil {
		panic(err)
	}
	if _, err := tx.Exec(`UPDATE posts SET category=? WHERE category=?`, 0, 1); err != nil {
		panic(err)
	}
}

func v16(tx *sql.Tx) {
	mustExec(tx, "ALTER TABLE comments ADD COLUMN `date_int` INT NOT NULL AFTER `date`")
	mustExec(tx, "UPDATE comments SET `date_int` = UNIX_TIMESTAMP(`date`)")
	mustExec(tx, "ALTER TABLE comments DROP COLUMN `date`")
	mustExec(tx, "ALTER TABLE comments CHANGE `date_int` `date` INT NOT NULL")

	mustExec(tx, "ALTER TABLE posts ADD COLUMN `date_int` INT NOT NULL AFTER `date`")
	mustExec(tx, "UPDATE posts SET `date_int` = UNIX_TIMESTAMP(`date`)")
	mustExec(tx, "ALTER TABLE posts DROP COLUMN `date`")
	mustExec(tx, "ALTER TABLE posts CHANGE `date_int` `date` INT NOT NULL")

	mustExec(tx, "ALTER TABLE posts ADD COLUMN `modified_int` INT NOT NULL AFTER `modified`")
	mustExec(tx, "UPDATE posts SET `modified_int` = UNIX_TIMESTAMP(`modified`)")
	mustExec(tx, "ALTER TABLE posts DROP COLUMN `modified`")
	mustExec(tx, "ALTER TABLE posts CHANGE `modified_int` `modified` INT NOT NULL")
}

// SQLite3 only

func v17(tx *sql.Tx) {
	mustExec(tx, `
	CREATE TABLE options2 (
		id INTEGER  PRIMARY KEY AUTOINCREMENT,
		name VARCHAR(64)  NOT NULL UNIQUE COLLATE NOCASE,
		value TEXT  NOT NULL
	)
	`)
	mustExec(tx, `INSERT INTO options2 SELECT * FROM options`)
	mustExec(tx, `DROP TABLE options`)
	mustExec(tx, `ALTER TABLE options2 RENAME TO options`)

	mustExec(tx, `
	CREATE TABLE tags2 (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE COLLATE NOCASE,
		alias INTEGER NOT NULL
	)
	`)
	mustExec(tx, `INSERT INTO tags2 SELECT * FROM tags`)
	mustExec(tx, `DROP TABLE tags`)
	mustExec(tx, `ALTER TABLE tags2 RENAME TO tags`)
}

func v18(tx *sql.Tx) {
	mustExec(tx, `CREATE TABLE IF NOT EXISTS pingbacks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at INTEGER  NOT NULL,
		post_id INTEGER  NOT NULL,
		title TEXT NOT NULL,
		source_url TEXT NOT NULL,
		UNIQUE (post_id, source_url)
	)`)
}
