package gateway

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
)

func (g *Gateway) GetComment(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	out := g.service.GetComment2(name)
	c.JSON(200, out)
}

func (g *Gateway) DeleteComment(c *gin.Context) {
	_ = utils.MustToInt64(c.Param("name"))
	commentName := utils.MustToInt64(c.Param("comment_name"))
	g.service.DeleteComment(nil, commentName)
}

// Comment ...
func (g *Gateway) listPostComments(c *gin.Context) {
	userCtx := g.auther.AuthCookie(c).Context(nil)

	name := utils.MustToInt64(c.Param("name"))
	limit := utils.MustToInt64(c.DefaultQuery("limit", "10"))
	offset := utils.MustToInt64(c.DefaultQuery("offset", "0"))

	comments, err := g.service.ListComments(userCtx, &protocols.ListCommentsRequest{
		Mode:    protocols.ListCommentsMode_ListCommentsModeTree,
		PostId:  name,
		Limit:   limit,
		Offset:  offset,
		OrderBy: "id DESC",
	})
	if err != nil {
		panic(err)
	}

	c.JSON(200, comments)
}

func (g *Gateway) createPostComment(c *gin.Context) {
	var comment protocols.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.Status(400)
		return
	}
	// TODO remove unwanted field values
	comment.Ip = c.ClientIP()
	comment.Date = datetime.ProtoLocal()
	user := g.auther.AuthCookie(c)
	cmt := g.service.CreateComment(user.Context(nil), &comment)
	c.JSON(200, cmt)
}
