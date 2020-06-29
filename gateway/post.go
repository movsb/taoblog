package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
)

// GetPost gets a post by its ID.
func (g *Gateway) GetPost(c *gin.Context) {
	id := utils.MustToInt64(c.Param("name"))
	p := g.service.GetPostByID(id)
	c.JSON(200, p)
}
