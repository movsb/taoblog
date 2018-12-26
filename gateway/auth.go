package gateway

import (
	"github.com/gin-gonic/gin"
)

func (g *Gateway) auth(c *gin.Context) {
	ok := g.auther.AuthCookie(c) || g.auther.AuthHeader(c)
	if !ok {
		c.Status(401)
		c.Abort()
	}
}
