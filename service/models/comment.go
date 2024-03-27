package models

import (
	"time"

	"github.com/movsb/taoblog/protocols"
	"github.com/xeonx/timeago"
)

// Comment in database.
type Comment struct {
	ID         int64  `json:"id"`
	Parent     int64  `json:"parent"`
	Root       int64  `json:"root"`
	PostID     int64  `json:"post_id"`
	Author     string `json:"author"`
	Email      string `json:"email"`
	URL        string `json:"url"`
	IP         string `json:"ip"`
	Date       int32  `json:"date"`
	SourceType string `json:"source_type"`
	Source     string `json:"source"`
	Content    string `json:"content"`
}

// TableName ...
func (Comment) TableName() string {
	return `comments`
}

// ToProtocols ...
func (c *Comment) ToProtocols(extra func(m *Comment, p *protocols.Comment)) *protocols.Comment {
	comment := protocols.Comment{
		Id:         c.ID,
		Parent:     c.Parent,
		Root:       c.Root,
		PostId:     c.PostID,
		Author:     c.Author,
		Url:        c.URL,
		Date:       c.Date,
		SourceType: c.SourceType,
		Source:     c.Source,
		Content:    c.Content,
		DateFuzzy:  timeago.Chinese.Format(time.Unix(int64(c.Date), 0)),
	}

	if extra != nil {
		extra(c, &comment)
	}

	return &comment
}

func In5min(t int32) bool {
	return time.Since(time.Unix(int64(t), 0)) < time.Minute*5
}

// Comments ...
type Comments []*Comment

// ToProtocols ...
func (cs Comments) ToProtocols(extra func(m *Comment, p *protocols.Comment)) []*protocols.Comment {
	comments := make([]*protocols.Comment, 0, len(cs))
	for _, comment := range cs {
		comments = append(comments, comment.ToProtocols(extra))
	}
	return comments
}
