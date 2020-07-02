package models

import (
	"strings"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
)

type Comment struct {
	ID         int64  `json:"id"`
	Parent     int64  `json:"parent"`
	Root       int64  `json:"root"`
	PostID     int64  `json:"post_id"`
	Author     string `json:"author"`
	Email      string `json:"email"`
	URL        string `json:"url"`
	IP         string `json:"ip"`
	Date       string `json:"date"`
	SourceType string `json:"source_type"`
	Source     string `json:"source"`
	Content    string `json:"content"`
}

// TableName ...
func (Comment) TableName() string {
	return `comments`
}

func (c *Comment) ToProtocols(adminEmail string, user *auth.User) *protocols.Comment {
	comment := protocols.Comment{
		Id:         c.ID,
		Parent:     c.Parent,
		Root:       c.Root,
		PostId:     c.PostID,
		Author:     c.Author,
		Url:        c.URL,
		Date:       datetime.My2Proto(c.Date),
		SourceType: c.SourceType,
		Source:     c.Source,
		Content:    c.Content,
		Avatar:     utils.Md5Str(c.Email),
		IsAdmin:    strings.EqualFold(c.Email, adminEmail),
	}

	if user.IsAdmin() {
		comment.Email = c.Email
		comment.Ip = c.IP
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
