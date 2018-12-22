package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) auth(c *gin.Context) {
	in := &protocols.AuthRequest{
		C: c,
	}
	out := g.server.Auth(in)
	if !out.Success {
		c.Status(401)
		c.Abort()
	}
}
