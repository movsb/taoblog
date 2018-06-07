package main

import (
	"database/sql"
	"fmt"
	"log"
)

type Comment struct {
	ID       int64
	Parent   int64
	Ancestor int64
	PostID   int64
	Author   string
	EMail    string
	URL      string
	IP       string
	Date     string
	Content  string
}

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

// DeleteComments deletes comment whose id is id
// It also deletes its children.
func (o *CommentManager) DeleteComments(id int64) error {
	query := fmt.Sprintf(`DELETE FROM comments WHERE id=%d OR ancestor=%d`, id, id)
	_, err := o.db.Exec(query)
	return err
}

// GetComment returns the specified comment object.
func (o *CommentManager) GetComment(id int64) (*Comment, error) {
	query := fmt.Sprintf(`SELECT id,parent,ancestor,post_id,author,email,url,ip,date,content FROM comments WHERE id=%d LIMIT 1`, id)
	row := o.db.QueryRow(query)
	cmt := &Comment{}
	err := row.Scan(&cmt.ID, &cmt.Parent, &cmt.Ancestor, &cmt.PostID, &cmt.Author, &cmt.EMail, &cmt.URL, &cmt.IP, &cmt.Date, &cmt.Content)
	return cmt, err
}

// GetRecentComments gets the recent comments
// TODO Not tested
func (o *CommentManager) GetRecentComments(num int) ([]*Comment, error) {
	var err error
	query := `SELECT id,parent,ancestor,post_id,author,email,url,ip,date,content FROM comments ORDER BY date DESC LIMIT ` + fmt.Sprint(num)
	rows, err := o.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cmts []*Comment

	for rows.Next() {
		var cmt Comment
		if err = rows.Scan(&cmt.ID, &cmt.Parent, &cmt.Ancestor, &cmt.PostID, &cmt.Author, &cmt.EMail, &cmt.URL, &cmt.IP, &cmt.Date, &cmt.Content); err != nil {
			return nil, err
		}
		cmts = append(cmts, &cmt)
	}

	return cmts, rows.Err()
}

// GetChildren gets all children comments of an ancestor
// TODO Not tested
func (o *CommentManager) GetChildren(id int64) ([]*Comment, error) {
	var err error

	query := `SELECT id,parent,ancestor,post_id,author,email,url,ip,date,content FROM comments WHERE ancestor=` + fmt.Sprint(id)
	rows, err := o.db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cmts []*Comment

	for rows.Next() {
		var cmt Comment
		if err = rows.Scan(&cmt.ID, &cmt.Parent, &cmt.Ancestor, &cmt.PostID, &cmt.Author, &cmt.EMail, &cmt.URL, &cmt.IP, &cmt.Date, &cmt.Content); err != nil {
			return nil, err
		}
		cmts = append(cmts, &cmt)
	}

	return cmts, rows.Err()
}
