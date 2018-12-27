package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service"
)

func (g *Gateway) GetPost(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	post := g.server.GetPost(name)
	c.JSON(200, post)
}

func (g *Gateway) ListPosts(c *gin.Context) {
	in := service.ListPostsRequest{
		Fields:  c.DefaultQuery("fields", "*"),
		Limit:   utils.MustToInt64(c.DefaultQuery("limit", "-1")),
		OrderBy: c.Query("order_by"),
	}
	posts := g.server.ListPosts(&in)
	c.JSON(200, posts)
}

func (g *Gateway) GetLatestPosts(c *gin.Context) {
	fields := c.DefaultQuery("fields", "*")
	limit := utils.MustToInt64(c.DefaultQuery("limit", "10"))
	posts := g.server.GetLatestPosts(fields, limit)
	c.JSON(200, posts)
}

func (g *Gateway) GetPostTitle(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	title := g.server.GetPostTitle(name)
	c.JSON(200, title)
}

func (g *Gateway) IncrementPostPageView(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	g.server.IncrementPostPageView(name)
	c.JSON(200, nil)
}

func (g *Gateway) GetPostCommentCount(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	count := g.server.GetPostCommentCount(name)
	c.JSON(200, count)
}
