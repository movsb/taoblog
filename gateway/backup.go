package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) GetBackup(c *gin.Context) {
	in := &protocols.GetBackupRequest{
		W: c.Writer,
	}
	g.server.GetBackup(in)
}
