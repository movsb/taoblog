package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/service"
)

type Gateway struct {
	router *gin.RouterGroup
	server *service.ImplServer
	auther *auth.Auth
}

func NewGateway(router *gin.RouterGroup, server *service.ImplServer, auther *auth.Auth) *Gateway {
	g := &Gateway{
		router: router,
		server: server,
		auther: auther,
	}

	g.routeOptions()
	g.routePosts()
	g.routeComments()
	g.routeTags()
	g.routeOthers()

	return g
}

func (g *Gateway) routeOthers() {
	g.router.GET("/avatar", g.GetAvatar)
	g.router.GET("/backup", g.auth, g.GetBackup)
}

func (g *Gateway) routeOptions() {
	c := g.router.Group("/options", g.auth)
	_ = c
	//c.GET("/:name", g.GetOption)
	//c.GET("", g.ListOptions)
}

func (g *Gateway) routePosts() {
	c := g.router.Group("/posts")
	g.router.GET("/posts!latest", g.GetLatestPosts)
	c.GET("/:name", g.GetPost)
	c.GET("", g.ListPosts)
	c.GET("/:name/title", g.GetPostTitle)
	c.POST("/:name/page_view", g.auth, g.IncrementPostPageView)
	c.GET("/:name/comments!count", g.GetPostCommentCount)
	// files
	c.GET("/:name/files/*file", g.GetFile)
	c.GET("/:name/files", g.ListFiles)
	c.POST("/:name/files/*file", g.auth, g.UploadFile)
	c.DELETE("/:name/files/*file", g.auth, g.DeleteFile)
}

func (g *Gateway) routeComments() {
	c := g.router.Group("/comments", g.auth)
	c.GET("/:name", g.GetComment)
	c.GET("", g.ListComments)
}

func (g *Gateway) routeTags() {
	c := g.router.Group("/tags")
	_ = c
	//g.router.GET("/tags!withCount", g.ListTagsWithCount)
}
