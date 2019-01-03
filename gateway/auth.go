package gateway

import (
	"github.com/gin-gonic/gin"
)

func (g *Gateway) auth(c *gin.Context) {
	cookieUser := g.auther.AuthCookie(c)
	headerUser := g.auther.AuthContext(c)
	authUser := g.auther.AuthHeader(c)
	if cookieUser.IsGuest() && headerUser.IsGuest() && authUser.IsGuest() {
		c.Status(401)
		c.Abort()
	}
}
