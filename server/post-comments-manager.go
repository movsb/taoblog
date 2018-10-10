package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/movsb/taoblog/server/modules/utils/datetime"
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

func (o *PostCommentsManager) GetPostComments(tx Querier, cid int64, offset int64, count int64, pid int64, ascent bool) ([]*AjaxComment, error) {
	cmts, err := cmtmgr.GetCommentAndItsChildren(tx, cid, offset, count, pid, ascent)
	if err != nil {
		return nil, err
	}

	adminEmail := strings.ToLower(optmgr.GetDef(tx, "email", ""))

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
