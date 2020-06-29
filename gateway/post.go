package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
)

// GetPost gets a post by its ID.
func (g *Gateway) GetPost(c *gin.Context) {
	id := utils.MustToInt64(c.Param("name"))
	p := g.service.GetPostByID(id)
	c.JSON(200, p)
}

func (g *Gateway) SetPostStatus(c *gin.Context) {
	p := protocols.Post{}
	if err := c.ShouldBindJSON(&p); err != nil {
		c.String(400, "%s", err)
		return
	}
	id := utils.MustToInt64(c.Param("name"))
	g.service.SetPostStatus(id, p.Status)
	c.Status(200)
}
