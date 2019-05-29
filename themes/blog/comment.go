package blog

import (
	"strings"

	"github.com/movsb/taoblog/modules/datetime"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// Comment ...
type Comment struct {
	*protocols.Comment
	PostTitle string
	IsAdmin   bool
}

func newComments(comments []*protocols.Comment, service *service.Service) []*Comment {
	cmts := []*Comment{}
	titles := make(map[int64]string)
	adminEmail := strings.ToLower(service.GetDefaultStringOption("email", ""))
	for _, c := range comments {
		title := ""
		if t, ok := titles[c.PostID]; ok {
			title = t
		} else {
			title = service.GetPostTitle(c.PostID)
			titles[c.PostID] = title
		}
		cmts = append(cmts, &Comment{
			Comment:   c,
			PostTitle: title,
			IsAdmin:   strings.ToLower(c.Email) == adminEmail,
		})
	}
	return cmts
}

// AuthorString ...
func (c *Comment) AuthorString() string {
	if c.IsAdmin {
		return "博主"
	}
	return c.Author
}

func (b *Blog) listPostComments(c *gin.Context) {
	userCtx := b.auth.AuthCookie(c).Context(nil)

	name := utils.MustToInt64(c.Param("name"))
	limit := utils.MustToInt64(c.DefaultQuery("limit", "10"))
	offset := utils.MustToInt64(c.DefaultQuery("offset", "0"))

	comments := b.service.ListComments(userCtx, &protocols.ListCommentsRequest{
		PostID:  name,
		Limit:   limit,
		Offset:  offset,
		OrderBy: "id DESC",
	})

	c.JSON(200, comments)
}

func (b *Blog) createPostComment(c *gin.Context) {
	comment := &protocols.Comment{
		PostID:  utils.MustToInt64(c.Param("name")),
		Parent:  utils.MustToInt64(c.DefaultPostForm("parent", "0")),
		Author:  c.DefaultPostForm("author", ""),
		Email:   c.DefaultPostForm("email", ""),
		URL:     c.DefaultPostForm("url", ""),
		IP:      c.ClientIP(),
		Date:    datetime.MyLocal(),
		Content: c.DefaultPostForm("content", ""),
	}
	user := b.auth.AuthCookie(c)
	comment = b.service.CreateComment(user.Context(nil), comment)
	c.JSON(200, comment)
}
