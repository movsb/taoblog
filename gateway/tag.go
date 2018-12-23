package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) ListTagsWithCount(c *gin.Context) {
	in := &protocols.ListTagsWithCountRequest{
		Limit:      toInt64(c.Query("limit")),
		MergeAlias: toInt64(c.Query("merge")) == 1,
	}
	out := g.server.ListTagsWithCount(in)
	c.JSON(200, out)
}
