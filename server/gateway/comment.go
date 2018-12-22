package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/server"
)

func (g *Gateway) ListComments(c *gin.Context) {
	parent := toInt64(c.Param("parent"))
	in := &server.ListCommentsRequest{
		Parent: parent,
	}
	out := g.server.ListComments(in)
	c.JSON(200, out.Comments)
}
