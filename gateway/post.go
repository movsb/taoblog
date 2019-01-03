package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
)

func (g *Gateway) GetPostCommentCount(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	count := g.server.GetPostCommentCount(name)
	c.JSON(200, count)
}
