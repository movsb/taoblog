package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) GetComment(c *gin.Context) {
	name := toInt64(c.Param("name"))
	in := &protocols.GetCommentRequest{
		Name: name,
	}
	out := g.server.GetComment(in)
	c.JSON(200, out)
}

func (g *Gateway) ListComments(c *gin.Context) {
	parent := toInt64(c.Query("parent"))
	in := &protocols.ListCommentsRequest{
		Parent: parent,
	}
	out := g.server.ListComments(in)
	c.JSON(200, out)
}
