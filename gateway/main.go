package gateway

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

type Gateway struct {
	router *gin.RouterGroup
	server protocols.IServer
}

func NewGateway(router *gin.RouterGroup, server protocols.IServer) *Gateway {
	g := &Gateway{
		router: router,
		server: server,
	}
	g.routeOptions()
	g.routePosts()
	g.routeComments()

	g.router.GET("/avatar", g.GetAvatar)
	g.router.GET("/backup", g.GetBackup)

	return g
}

func (g *Gateway) routeOptions() {
	c := g.router.Group("/options")
	c.GET("/:name", g.GetOption)
	c.GET("", g.ListOptions)
}

func (g *Gateway) routePosts() {
	c := g.router.Group("/posts")
	c.GET("/:name", g.GetPost)
	c.GET("", g.ListPosts)
}

func (g *Gateway) routeComments() {
	c := g.router.Group("/comments")
	c.GET("/:name", g.GetComment)
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
