package main

import (
	"database/sql"
	"fmt"
	"log"
)

type Comment struct {
	ID       int64      `json:"id"`
	Parent   int64      `json:"parent"`
	Ancestor int64      `json:"ancestor"`
	PostID   int64      `json:"post_id"`
	Author   string     `json:"author"`
	EMail    string     `json:"email"`
	URL      string     `json:"url"`
	IP       string     `json:"ip"`
	Date     string     `json:"date"`
	Content  string     `json:"content"`
	Children []*Comment `json:"children"`
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

	cmts := make([]*Comment, 0)

	for rows.Next() {
		var cmt Comment
		if err = rows.Scan(&cmt.ID, &cmt.Parent, &cmt.Ancestor, &cmt.PostID, &cmt.Author, &cmt.EMail, &cmt.URL, &cmt.IP, &cmt.Date, &cmt.Content); err != nil {
			return nil, err
		}
		cmts = append(cmts, &cmt)
	}

	return cmts, rows.Err()
}

// GetAncestor returns the ancestor of a comment
func (o *CommentManager) GetAncestor(id int64, returnIDIfZero bool) (int64, error) {
	query := `SELECT ancestor FROM comments WHERE id=` + fmt.Sprint(id) + ` LIMIT 1`
	row := o.db.QueryRow(query)
	var aid int64
	if err := row.Scan(&aid); err != nil {
		return -1, err
	}

	if aid != 0 {
		return aid, nil
	}

	if returnIDIfZero {
		return id, nil
	}

	return 0, nil
}

func (o *CommentManager) GetCommentAndItsChildren(cid int64, offset int64, count int64, pid int64, ascent bool) ([]*Comment, error) {
	var query string

	if cid > 0 {
		query += `SELECT id,parent,ancestor,post_id,author,email,url,ip,date,content FROM comments WHERE id=` + fmt.Sprint(cid)
	} else {
		query += `SELECT id,parent,ancestor,post_id,author,email,url,ip,date,content FROM comments WHERE parent=0`
		if pid > 0 {
			query += ` AND post_id=` + fmt.Sprint(pid)
		}

		if ascent {
			query += ` ORDER BY id ASC`
		} else {
			query += ` ORDER BY id DESC`
		}

		if count > 0 {
			if offset >= 0 {
				query += fmt.Sprintf(" LIMIT %d,%d", offset, count)
			} else {
				query += fmt.Sprintf(" LIMIT %d", count)
			}
		}
	}

	cmts := make([]*Comment, 0)

	rows, err := o.db.Query(query)
	if err != nil {
		return cmts, err
	}

	defer rows.Close()

	for rows.Next() {
		var cmt Comment
		if err = rows.Scan(&cmt.ID, &cmt.Parent, &cmt.Ancestor, &cmt.PostID, &cmt.Author, &cmt.EMail, &cmt.URL, &cmt.IP, &cmt.Date, &cmt.Content); err != nil {
			return cmts, err
		}
		cmts = append(cmts, &cmt)
	}

	for _, cmt := range cmts {
		cmt.Children, err = o.GetChildren(cmt.ID)
		if err != nil {
			return nil, err
		}
	}

	return cmts, rows.Err()
}
