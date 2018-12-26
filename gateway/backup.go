package gateway

import (
	"github.com/gin-gonic/gin"
)

func (g *Gateway) GetBackup(c *gin.Context) {
	g.server.GetBackup(c.Writer)
}
