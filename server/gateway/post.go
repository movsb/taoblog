package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/server"
)

func (g *Gateway) GetPost(c *gin.Context) {
	name := toInt64(c.Param("name"))
	in := &server.GetPostRequest{
		Name: name,
	}
	out := g.server.GetPost(in)
	c.JSON(200, out)
}

func (g *Gateway) ListPosts(c *gin.Context) {
	in := &server.ListPostsRequest{}
	out := g.server.ListPosts(in)
	c.JSON(200, out)
}
