package data

import (
	"context"

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

// ListLatestComments ...
func (d *Data) ListLatestComments() (posts []*LatestCommentsByPost) {
	comments, err := d.svc.ListComments(d.User.Context(context.TODO()),
		&protocols.ListCommentsRequest{
			Mode:    protocols.ListCommentsRequest_Flat,
			Limit:   15,
			OrderBy: "date DESC",
		})
	if err != nil {
		panic(err)
	}
	postsMap := make(map[int64]*LatestCommentsByPost)
	for _, c := range comments.Comments {
		p, ok := postsMap[c.PostId]
		if !ok {
			p = &LatestCommentsByPost{
				PostID:    c.PostId,
				PostTitle: d.svc.GetPostTitle(c.PostId),
			}
			postsMap[c.PostId] = p
			posts = append(posts, p)
		}
		p.Comments = append(p.Comments, &Comment{
			Comment: c,
		})
	}
	return
}
