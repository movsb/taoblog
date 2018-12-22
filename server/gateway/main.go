package gateway

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/server"
)

type Gateway struct {
	router *gin.RouterGroup
	server server.IServer
}

func NewGateway(router *gin.RouterGroup, server server.IServer) *Gateway {
	g := &Gateway{
		router: router,
		server: server,
	}
	g.routeComments()
	return g
}

func (g *Gateway) routeComments() {
	c := g.router.Group("/comments")
	c.GET("", g.ListComments)
}

// TODO remove
func toInt64(s string) int64 {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		//panic(fmt.Errorf("expect number: %s", s))
	}
	return n
}
