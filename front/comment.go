package front

import (
	"html/template"

	"github.com/movsb/taoblog/protocols"
)

type Comment struct {
	*protocols.Comment
	server protocols.IServer
}

func newComment(comment *protocols.Comment, server protocols.IServer) *Comment {
	return &Comment{
		Comment: comment,
		server:  server,
	}
}

func newComments(comments []*protocols.Comment, server protocols.IServer) []*Comment {
	cmts := []*Comment{}
	for _, c := range comments {
		cmts = append(cmts, newComment(c, server))
	}
	return cmts
}

func (c *Comment) AuthorString() string {
	mail := c.server.GetOption(&protocols.GetOptionRequest{
		Name:    "email",
		Default: true,
		Value:   "",
	}).Value
	if mail == c.Email {
		return "博主"
	}
	return c.Author
}

func (c *Comment) PostTitle() template.HTML {
	return template.HTML(c.server.GetPostTitle(c.PostID))
}
