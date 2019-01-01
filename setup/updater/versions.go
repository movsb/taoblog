package main

import (
	"database/sql"
	"encoding/json"

	"github.com/movsb/taoblog/modules/taorm"
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
	if err := taorm.QueryRows(&metas, tx, `SELECT * FROM post_metas WHERE type='post' OR type='page'`); err != nil {
		panic(err)
	}
	for _, meta := range metas {
		var pm v3Meta
		query := `SELECT id,metas FROM posts WHERE id=?`
		if err := taorm.QueryRows(&pm, tx, query, meta.Tid); err != nil {
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
