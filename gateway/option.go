package gateway

import (
	"database/sql"

	"github.com/movsb/taoblog/protocols"

	"github.com/gin-gonic/gin"
)

func (g *Gateway) GetOption(c *gin.Context) {
	name := c.Param("name")
	option, err := g.service.GetOption(name)
	if err == sql.ErrNoRows {
		c.Status(404)
		return
	}
	if err != nil {
		c.Status(500)
		return
	}
	c.JSON(200, option)
}

func (g *Gateway) SetOption(c *gin.Context) {
	name := c.Param("name")
	option := protocols.Option{}
	if err := c.ShouldBindJSON(&option); err != nil {
		c.Status(400)
		return
	}
	g.service.SetOption(name, option.Value)
}
