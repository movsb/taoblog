package data

import (
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// Comment ...
type Comment struct {
	*protocols.Comment
	PostTitle string
	Content   string
}

func newComments(comments []*protocols.Comment, service *service.Service) []*Comment {
	cmts := []*Comment{}
	titles := make(map[int64]string)
	for _, c := range comments {
		title := ""
		if t, ok := titles[c.PostId]; ok {
			title = t
		} else {
			title = service.GetPostTitle(c.PostId)
			titles[c.PostId] = title
		}
		content := c.Source
		if content == "" {
			content = c.Content
		}
		cmts = append(cmts, &Comment{
			Comment:   c,
			PostTitle: title,
			Content:   content,
		})
	}
	return cmts
}
