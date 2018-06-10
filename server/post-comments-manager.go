package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"strings"
)

type CommentPrivate struct {
	ID       int64             `json:"id"`
	Parent   int64             `json:"parent"`
	Ancestor int64             `json:"ancestor"`
	PostID   int64             `json:"post_id"`
	Author   string            `json:"author"`
	EMail    string            `json:"email"`
	URL      string            `json:"url"`
	IP       string            `json:"ip"`
	Date     string            `json:"date"`
	Content  string            `json:"content"`
	Children []*CommentPrivate `json:"children"`

	Avatar  string `json:"avatar"`
	IsAdmin bool   `json:"is_admin"`
}

type CommentPublic struct {
	ID       int64            `json:"id"`
	Parent   int64            `json:"parent"`
	Ancestor int64            `json:"ancestor"`
	PostID   int64            `json:"post_id"`
	Author   string           `json:"author"`
	URL      string           `json:"url"`
	Date     string           `json:"date"`
	Content  string           `json:"content"`
	Children []*CommentPublic `json:"children"`

	Avatar  string `json:"avatar"`
	IsAdmin bool   `json:"is_admin"`
}

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

func (o *PostCommentsManager) GetPostCommentsPublic(cid int64, offset int64, count int64, pid int64, ascent bool) ([]*CommentPublic, error) {
	cmts, err := cmtmgr.GetCommentAndItsChildren(cid, offset, count, pid, ascent)
	if err != nil {
		return nil, err
	}

	adminEmail := strings.ToLower(optmgr.GetDef("email", ""))

	md5it := func(s string) string {
		md5 := md5.New()
		md5.Write([]byte(s))
		return fmt.Sprintf("%x", md5.Sum(nil))
	}

	convert := func(c *Comment) *CommentPublic {
		pc := &CommentPublic{
			ID:       c.ID,
			Parent:   c.Parent,
			Ancestor: c.Ancestor,
			PostID:   c.PostID,
			Author:   c.Author,
			URL:      c.URL,
			Date:     c.Date,
			Content:  c.Content,
		}

		pc.Avatar = md5it(c.EMail)
		pc.IsAdmin = strings.ToLower(c.EMail) == adminEmail

		return pc
	}

	pcmts := make([]*CommentPublic, 0, len(cmts))

	for _, c := range cmts {
		pc := convert(c)
		for _, cc := range c.Children {
			pc.Children = append(pc.Children, convert(cc))
		}
		if len(pc.Children) == 0 {
			pc.Children = make([]*CommentPublic, 0)
		}
		pcmts = append(pcmts, pc)
	}

	return pcmts, nil
}

func (o *PostCommentsManager) GetPostCommentsPrivate(cid int64, offset int64, count int64, pid int64, ascent bool) ([]*CommentPrivate, error) {
	cmts, err := cmtmgr.GetCommentAndItsChildren(cid, offset, count, pid, ascent)
	if err != nil {
		return nil, err
	}

	adminEmail := strings.ToLower(optmgr.GetDef("email", ""))

	md5it := func(s string) string {
		md5 := md5.New()
		md5.Write([]byte(s))
		return fmt.Sprintf("%x", md5.Sum(nil))
	}

	convert := func(c *Comment) *CommentPrivate {
		pc := &CommentPrivate{
			ID:       c.ID,
			Parent:   c.Parent,
			Ancestor: c.Ancestor,
			PostID:   c.PostID,
			Author:   c.Author,
			EMail:    c.EMail,
			URL:      c.URL,
			IP:       c.IP,
			Date:     c.Date,
			Content:  c.Content,
		}

		pc.Avatar = md5it(c.EMail)
		pc.IsAdmin = strings.ToLower(c.EMail) == adminEmail

		return pc
	}

	pcmts := make([]*CommentPrivate, 0, len(cmts))

	for _, c := range cmts {
		pc := convert(c)
		for _, cc := range c.Children {
			pc.Children = append(pc.Children, convert(cc))
		}
		if len(pc.Children) == 0 {
			pc.Children = make([]*CommentPrivate, 0)
		}
		pcmts = append(pcmts, pc)
	}

	return pcmts, nil
}
