package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) createPostComment(c *gin.Context) {
	var comment protocols.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.Status(400)
		return
	}
	// TODO remove unwanted field values
	comment.Ip = c.ClientIP()
	comment.Date = datetime.ProtoLocal()
	user := g.auther.AuthCookie(c)
	cmt := g.service.CreateComment(user.Context(nil), &comment)
	c.JSON(200, cmt)
}
