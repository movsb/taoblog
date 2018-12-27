package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service"
)

func (g *Gateway) GetComment(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	out := g.server.GetComment(name)
	c.JSON(200, out)
}

func (g *Gateway) ListComments(c *gin.Context) {
	//parent := toInt64(c.Query("parent"))
	in := &service.ListCommentsRequest{}
	out := g.server.ListComments(in)
	c.JSON(200, out)
}
