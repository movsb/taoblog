package data

import (
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/modules/renderers"
)

// Comment ...
type Comment struct {
	*protocols.Comment
}

// Text ...
func (c *Comment) Text() string {
	if c.Source != `` {
		return c.Source
	}
	return c.Content
}

func (c *Comment) PrettyText() string {
	prettifier := renderers.Prettifier{}
	text, err := prettifier.Prettify(c.Content)
	if err != nil {
		return c.Text()
	}
	return text
}

// LatestCommentsByPost ...
type LatestCommentsByPost struct {
	PostTitle string
	PostID    int64
	Comments  []*Comment
}
