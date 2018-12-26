package front

import (
	"html/template"

	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

type Comment struct {
	*models.Comment
	server *service.ImplServer
}

func newComment(comment *models.Comment, server *service.ImplServer) *Comment {
	return &Comment{
		Comment: comment,
		server:  server,
	}
}

func newComments(comments []*models.Comment, server *service.ImplServer) []*Comment {
	cmts := []*Comment{}
	for _, c := range comments {
		cmts = append(cmts, newComment(c, server))
	}
	return cmts
}

func (c *Comment) AuthorString() string {
	mail := c.server.GetDefaultStringOption("email", "")
	if mail == c.Email {
		return "博主"
	}
	return c.Author
}

func (c *Comment) PostTitle() template.HTML {
	return template.HTML(c.server.GetPostTitle(c.PostID))
}
