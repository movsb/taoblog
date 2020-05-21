package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) GetAvatar(c *gin.Context) {
	query := c.Request.URL.RawQuery
	in := &protocols.GetAvatarRequest{
		Query:           c.Param(`hash`) + `?` + query,
		IfModifiedSince: c.GetHeader("If-Modified-Since"),
		IfNoneMatch:     c.GetHeader("If-None-Match"),
		SetStatus: func(statusCode int) {
			c.Status(statusCode)
		},
		SetHeader: func(name string, value string) {
			c.Header(name, value)
		},
		W: c.Writer,
	}
	g.service.GetAvatar(in)
}
