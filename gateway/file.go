package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
)

func (g *Gateway) GetFile(c *gin.Context) {
	postID := utils.MustToInt64(c.Param("name"))
	file := c.Param("file")
	fp := g.service.GetFile(postID, file)
	c.File(fp)
}

func (g *Gateway) ListFiles(c *gin.Context) {
	postID := utils.MustToInt64(c.Param("name"))
	files, err := g.service.ListFiles(postID)
	if err != nil {
		c.String(500, "%v", err)
		return
	}
	c.JSON(200, files)
}

func (g *Gateway) CreateFile(c *gin.Context) {
	postID := utils.MustToInt64(c.Param("name"))
	file := c.Param("file")
	if err := g.service.CreateFile(postID, file, c.Request.Body); err != nil {
		c.JSON(500, err)
		return
	}
	c.Status(200)
}

func (g *Gateway) DeleteFile(c *gin.Context) {
	postID := utils.MustToInt64(c.Param("name"))
	file := c.Param("file")
	if err := g.service.DeleteFile(postID, file); err != nil {
		c.JSON(500, err)
		return
	}
	c.Status(200)
}
