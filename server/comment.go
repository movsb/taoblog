package main

import (
	"database/sql"
	"fmt"
	"log"
)

type CommentManager struct {
	db *sql.DB
}

func newCommentManager(db *sql.DB) *CommentManager {
	return &CommentManager{
		db: db,
	}
}

func (o *CommentManager) GetAllCount() (count uint) {
	query := `SELECT count(*) as size FROM comments`
	row := o.db.QueryRow(query)
	err := row.Scan(&count)
	if err != nil {
		log.Println(err)
	}
	return
}

func (o *CommentManager) SetStatus(id int64, st string) error {
	query := `UPDATE comments SET status=? WHERE id=` + fmt.Sprint(id) + ` LIMIT 1`
	ret, err := o.db.Exec(query, st)
	if err != nil {
		return err
	}

	_ = ret

	/*
		if rows, err := ret.RowsAffected(); err != nil || rows == 1 {
			return nil
		} else {
			return fmt.Errorf("设置失败：rows:%d, err:%s", rows, err)
		}
	*/

	return nil
}
