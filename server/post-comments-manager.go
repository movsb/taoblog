package main

import (
	"database/sql"
	"fmt"
)

type PostCommentsManager struct {
	db *sql.DB
}

func newPostCommentsManager(db *sql.DB) *PostCommentsManager {
	return &PostCommentsManager{
		db: db,
	}
}

func (o *PostCommentsManager) UpdatePostCommentsCount(pid int64) error {
	sql := `UPDATE posts INNER JOIN (SELECT post_id,count(post_id) count FROM comments WHERE post_id=%d) x ON posts.id=x.post_id SET posts.comments=x.count WHERE posts.id=%d`
	sql = fmt.Sprintf(sql, pid, pid)
	_, err := o.db.Exec(sql)
	return err
}

func (o *PostCommentsManager) DeletePostComment(cid int64) error {
	var err error
	cmt, err := cmtmgr.GetComment(cid)
	if err != nil {
		return err
	}

	err = cmtmgr.DeleteComments(cid)
	if err != nil {
		return err
	}

	err = o.UpdatePostCommentsCount(cmt.PostID)
	if err != nil {
		return err
	}

	return nil
}
