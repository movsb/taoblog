package main

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"./internal/utils/datetime"
)

type AjaxComment struct {
	c        *Comment
	Children []*AjaxComment
	Avatar   string
	IsAdmin  bool
	private  bool
}

func (o *AjaxComment) marshal(ac *AjaxComment, sb *bytes.Buffer, private bool, comma bool) error {
	var err error

	c := ac.c

	f := func(name string, v interface{}) {
		if err != nil {
			return
		}

		var by []byte

		by, err = json.Marshal(v)
		if err != nil {
			return
		}

		_, err = sb.WriteString(
			fmt.Sprintf(`"%s":%s,`, name, string(by)),
		)
	}

	sb.WriteString(`{`)

	f(`id`, c.ID)
	f(`parent`, c.Parent)
	f(`ancestor`, c.Ancestor)
	f(`post_id`, c.PostID)
	f(`author`, c.Author)
	f(`url`, c.URL)
	f(`date`, c.Date)
	f(`content`, c.Content)

	f(`avatar`, ac.Avatar)
	f(`is_admin`, ac.IsAdmin)

	if private {
		f(`email`, c.EMail)
		f(`ip`, c.IP)
	}

	_, err = sb.WriteString(fmt.Sprint(`"children":[`))

	for i, cc := range ac.Children {
		comma2 := i != len(ac.Children)-1
		err = o.marshal(cc, sb, o.private, comma2)
	}

	_, err = sb.WriteString(`]`) // no comma

	_, err = sb.WriteString(`}`)

	if comma {
		sb.WriteByte(',')
	}

	return err
}

func (o *AjaxComment) MarshalJSON() ([]byte, error) {
	var sb = &bytes.Buffer{}

	var err error

	err = o.marshal(o, sb, o.private, false)
	if err != nil {
		return nil, err
	}

	return sb.Bytes(), nil
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

func (o *PostCommentsManager) GetPostComments(offset int64, count int64, pid int64, ascent bool) ([]*AjaxComment, error) {
	cmts, err := cmtmgr.GetCommentAndItsChildren(0, offset, count, pid, ascent)
	if err != nil {
		return nil, err
	}

	adminEmail := strings.ToLower(optmgr.GetDef("email", ""))

	md5it := func(s string) string {
		md5 := md5.New()
		md5.Write([]byte(s))
		return fmt.Sprintf("%x", md5.Sum(nil))
	}

	convert := func(c *Comment) *AjaxComment {
		pc := &AjaxComment{
			c: c,
		}

		pc.c.Date = datetime.My2Local(pc.c.Date)

		pc.Avatar = md5it(c.EMail)
		pc.IsAdmin = strings.ToLower(c.EMail) == adminEmail

		return pc
	}

	pcmts := make([]*AjaxComment, 0, len(cmts))

	for _, c := range cmts {
		pc := convert(c)
		for _, cc := range c.Children {
			pc.Children = append(pc.Children, convert(cc))
		}
		if len(pc.Children) == 0 {
			pc.Children = make([]*AjaxComment, 0)
		}
		pcmts = append(pcmts, pc)
	}

	return pcmts, nil
}
