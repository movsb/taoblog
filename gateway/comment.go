package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) GetComment(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	out := g.server.GetComment(name)
	c.JSON(200, out)
}

func (g *Gateway) ListComments(c *gin.Context) {
	//parent := toInt64(c.Query("parent"))
	in := &protocols.ListCommentsRequest{}
	out := g.server.ListComments(in)
	c.JSON(200, out)
}

func (g *Gateway) DeleteComment(c *gin.Context) {
	_ = utils.MustToInt64(c.Param("name"))
	commentName := utils.MustToInt64(c.Param("comment_name"))
	g.server.DeleteComment(nil, commentName)
}
