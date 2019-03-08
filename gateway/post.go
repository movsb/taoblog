package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) GetPostCommentCount(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	count := g.service.GetPostCommentCount(name)
	c.JSON(200, count)
}

func (g *Gateway) CreatePost(c *gin.Context) {
	p := protocols.Post{}
	if err := c.ShouldBindJSON(&p); err != nil {
		c.String(400, "%s", err)
		return
	}
	g.service.CreatePost(&p)
	c.JSON(200, &p)
}

func (g *Gateway) UpdatePost(c *gin.Context) {
	p := protocols.Post{}
	if err := c.ShouldBindJSON(&p); err != nil {
		c.String(400, "%s", err)
		return
	}
	p.ID = utils.MustToInt64(c.Param("name"))
	g.service.UpdatePost(&p)
	c.JSON(200, &p)
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
