package gateway

import (
	"github.com/gin-gonic/gin"
)

func (g *Gateway) GetBackup(c *gin.Context) {
	g.service.GetBackup(c.Writer)
}
