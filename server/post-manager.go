package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"./internal/post_translators"
)

type xPostManager struct {
	db *sql.DB
}

func newPostManager(db *sql.DB) *xPostManager {
	return &xPostManager{
		db: db,
	}
}

// returns nil if post presents
func (me *xPostManager) has(id int64) error {
	query := `SELECT id FROM posts WHERE id=` + strconv.FormatInt(id, 10)
	rows := me.db.QueryRow(query)
	pid := 0
	err := rows.Scan(&pid)
	return err
}

func (me *xPostManager) update(id int64, typ string, source string) error {
	var tr post_translators.PostTranslator
	var content string
	var err error

	switch typ {
	case "html":
		tr = &post_translators.HTMLTranslator{}
	case "markdown":
		tr = &post_translators.MarkdownTranslator{}
	}

	if tr == nil {
		return errors.New("no translator found for " + typ)
	}

	content, err = tr.Translate(source)
	if err != nil {
		return err
	}

	modTime := time.Now().UTC().Format("2006:01:02 15:04:05")

	ret, err := me.db.Exec(
		"UPDATE posts SET content=?,source=?,source_type=?,modified=? WHERE id=? LIMIT 1",
		content,
		source,
		typ,
		modTime,
		id,
	)

	if err != nil {
		return err
	}

	/*
		if n, err := ret.RowsAffected(); err != nil || n != 1 {
			return errors.New("affected rows != 1: n=" + strconv.FormatInt(n, 10))
		}
	*/

	_ = ret

	return nil
}

func (me *xPostManager) getCommentCount(pid int64) (count int) {
	query := `SELECT comments FROM posts WHERE id=` + fmt.Sprint(pid) + ` LIMIT 1`
	row := me.db.QueryRow(query)
	switch row.Scan(&count) {
	case sql.ErrNoRows:
		count = -1
	case nil:
		break
	default:
		count = -1
	}
	return
}
