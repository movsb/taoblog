package models

import (
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type Comment struct {
	ID         int64  `json:"id"`
	Parent     int64  `json:"parent"`
	Root       int64  `json:"root"`
	PostID     int64  `json:"post_id"`
	Author     string `json:"author"`
	Email      string `json:"email"`
	UserID     int32  `json:"user_id"`
	URL        string `json:"url"`
	IP         string `json:"ip"`
	Date       int32  `json:"date"`
	Modified   int32  `json:"modified"`
	SourceType string `json:"source_type"`
	Source     string `json:"source"`

	DateTimezone     string
	ModifiedTimezone string
}

func (Comment) TableName() string {
	return `comments`
}

// 以下字段由 setCommentExtraFields 提供/清除。
// - Email
// - Ip
// - GeoLocation
// - CanEdit
// - Avatar
// - Content
func (c *Comment) ToProto(redact func(c *proto.Comment)) *proto.Comment {
	comment := proto.Comment{
		Id:         c.ID,
		Parent:     c.Parent,
		Root:       c.Root,
		PostId:     c.PostID,
		Author:     c.Author,
		Email:      c.Email,
		UserId:     c.UserID,
		Ip:         c.IP,
		Url:        c.URL,
		Date:       c.Date,
		Modified:   c.Modified,
		SourceType: c.SourceType,
		Source:     c.Source,

		// TODO 用评论自带时区。
		DateFuzzy: utils.RelativeDate(time.Unix(int64(c.Date), 0)),

		DateTimezone:     c.DateTimezone,
		ModifiedTimezone: c.ModifiedTimezone,
	}
	redact(&comment)
	return &comment
}

type Comments []*Comment

func (cs Comments) ToProto(redact func(c *proto.Comment)) []*proto.Comment {
	comments := make([]*proto.Comment, 0, len(cs))
	for _, comment := range cs {
		comments = append(comments, comment.ToProto(redact))
	}
	return comments
}
