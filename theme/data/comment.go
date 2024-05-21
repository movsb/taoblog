package data

import (
	"github.com/movsb/taoblog/protocols"
)

// Comment ...
type Comment struct {
	*proto.Comment
}

// Text ...
func (c *Comment) Text() string {
	if c.Source != `` {
		return c.Source
	}
	return c.Content
}

// LatestCommentsByPost ...
type LatestCommentsByPost struct {
	Post     *Post
	Comments []*Comment
}
