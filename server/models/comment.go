package models

import (
	"github.com/movsb/taoblog/modules/server"
)

// Comment ...
type Comment struct {
	ID       int64
	Parent   int64
	Ancestor int64
	PostID   int64
	Author   string
	Email    string
	URL      string
	IP       string
	Date     string
	Content  string
}

func (c *Comment) Serialize() *server.Comment {
	sc := &server.Comment{
		ID:       c.ID,
		Parent:   c.Parent,
		Ancestor: c.Ancestor,
		PostID:   c.PostID,
		Author:   c.Author,
		Email:    c.Email,
		URL:      c.URL,
		IP:       c.IP,
		Date:     c.Date,
		Content:  c.Content,
	}
	return sc
}

type Comments []*Comment

func (cs Comments) Serialize() []*server.Comment {
	sc := []*server.Comment{}
	for _, c := range cs {
		sc = append(sc, c.Serialize())
	}
	return sc
}
