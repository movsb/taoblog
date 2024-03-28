package data

import (
	"context"

	"github.com/movsb/taoblog/protocols"
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
			Mode:    protocols.ListCommentsMode_ListCommentsModeFlat,
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
