package migration

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/movsb/taoblog/auth"

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
	parts := strings.Split(login, ",")
	if len(parts) != 2 {
		panic("invalid login value")
	}
	savedAuth := auth.SavedAuth{
		Username: parts[0],
		Password: parts[1],
	}
	login = savedAuth.Encode()
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
