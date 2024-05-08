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
	Modified   int32  `json:"modified"`
	SourceType string `json:"source_type"`
	Source     string `json:"source"`
}

// TableName ...
func (Comment) TableName() string {
	return `comments`
}

// ToProtocols ...
// 以下字段由 setCommentExtraFields 提供/清除。
// - IsAdmin
// - Email
// - Ip
// - GeoLocation
// - CanEdit
// - Avatar
// - Content
func (c *Comment) ToProtocols(redact func(c *protocols.Comment)) *protocols.Comment {
	comment := protocols.Comment{
		Id:         c.ID,
		Parent:     c.Parent,
		Root:       c.Root,
		PostId:     c.PostID,
		Author:     c.Author,
		Email:      c.Email,
		Ip:         c.IP,
		Url:        c.URL,
		Date:       c.Date,
		Modified:   c.Modified,
		SourceType: c.SourceType,
		Source:     c.Source,
		DateFuzzy:  timeago.Chinese.Format(time.Unix(int64(c.Date), 0)),
	}
	redact(&comment)
	return &comment
}

// Comments ...
type Comments []*Comment

// ToProtocols ...
func (cs Comments) ToProtocols(redact func(c *protocols.Comment)) []*protocols.Comment {
	comments := make([]*protocols.Comment, 0, len(cs))
	for _, comment := range cs {
		comments = append(comments, comment.ToProtocols(redact))
	}
	return comments
}
