package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) GetOption(c *gin.Context) {
	name := c.Param("name")
	in := &protocols.GetOptionRequest{
		Name: name,
	}
	out := g.server.GetOption(in)
	c.JSON(200, out)
}

func (g *Gateway) ListOptions(c *gin.Context) {
	in := &protocols.ListOptionsRequest{}
	out := g.server.ListOptions(in)
	c.JSON(200, out)
}
