package main

import (
	"fmt"
)

type PostCommentsManager struct {
}

func newPostCommentsManager() *PostCommentsManager {
	return &PostCommentsManager{}
}

func (o *PostCommentsManager) UpdatePostCommentsCount(tx Querier, pid int64) error {
	query := `UPDATE posts INNER JOIN (SELECT count(post_id) count FROM comments WHERE post_id=%d) x ON posts.id=%d SET posts.comments=x.count`
	query = fmt.Sprintf(query, pid, pid)
	fmt.Println(query)
	_, err := tx.Exec(query)
	return err
}

func (o *PostCommentsManager) DeletePostComment(tx Querier, cid int64) error {
	var err error
	cmt, err := cmtmgr.GetComment(tx, cid)
	if err != nil {
		return err
	}

	err = cmtmgr.DeleteComments(tx, cid)
	if err != nil {
		return err
	}

	err = o.UpdatePostCommentsCount(tx, cmt.PostID)
	if err != nil {
		return err
	}

	return nil
}
