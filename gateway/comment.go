package gateway

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) GetComment(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	out := g.service.GetComment2(name)
	c.JSON(200, out)
}

func (g *Gateway) DeleteComment(c *gin.Context) {
	_ = utils.MustToInt64(c.Param("name"))
	commentName := utils.MustToInt64(c.Param("comment_name"))
	g.service.DeleteComment(nil, commentName)
}

// SetCommentPostID ...
func (g *Gateway) SetCommentPostID(c *gin.Context) {
	cmt := protocols.Comment{}
	if err := c.ShouldBindJSON(&cmt); err != nil {
		c.String(400, `%s`, err)
		return
	}
	g.service.SetCommentPostID(context.TODO(), cmt.Id, cmt.PostId)
}
