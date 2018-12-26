package gateway

import (
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/service"
)

func (g *Gateway) GetAvatar(c *gin.Context) {
	query := c.Request.URL.RawQuery
	query, _ = url.QueryUnescape(query)
	ifModified := c.GetHeader("If-Modified-Since")
	ifNoneMatch := c.GetHeader("If-None-Match")
	in := &service.GetAvatarRequest{
		Query:           query,
		IfModifiedSince: ifModified,
		IfNoneMatch:     ifNoneMatch,
		SetStatus: func(statusCode int) {
			c.Status(statusCode)
		},
		SetHeader: func(name string, value string) {
			c.Header(name, value)
		},
		W: c.Writer,
	}
	g.server.GetAvatar(in)
}
