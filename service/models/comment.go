package models

import (
	"strings"

	"github.com/movsb/taoblog/auth"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
)

type Comment struct {
	ID       int64  `json:"id"`
	Parent   int64  `json:"parent"`
	Ancestor int64  `json:"ancestor"`
	PostID   int64  `json:"post_id"`
	Author   string `json:"author"`
	Email    string `json:"email"`
	URL      string `json:"url"`
	IP       string `json:"ip"`
	Date     string `json:"date"`
	Content  string `json:"content"`
}

func (c *Comment) ToProtocols(adminEmail string, user *auth.User) *protocols.Comment {
	comment := protocols.Comment{
		ID:       c.ID,
		Parent:   c.Parent,
		Ancestor: c.Ancestor,
		PostID:   c.PostID,
		Author:   c.Author,
		URL:      c.URL,
		Date:     c.Date,
		Content:  c.Content,
	}

	comment.Avatar = utils.Md5Str(c.Email)
	comment.IsAdmin = strings.EqualFold(c.Email, adminEmail)

	if user.IsAdmin() {
		comment.Email = c.Email
		comment.IP = c.IP
	}

	return &comment
}

type Comments []*Comment

func (cs Comments) ToProtocols(adminEmail string, user *auth.User) []*protocols.Comment {
	comments := make([]*protocols.Comment, 0, len(cs))
	for _, comment := range cs {
		comments = append(comments, comment.ToProtocols(adminEmail, user))
	}
	return comments
}
