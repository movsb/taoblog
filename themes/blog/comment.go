package blog

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// Comment ...
type Comment struct {
	*protocols.Comment
	PostTitle string
	Content   string
}

func newComments(comments []*protocols.Comment, service *service.Service) []*Comment {
	cmts := []*Comment{}
	titles := make(map[int64]string)
	for _, c := range comments {
		title := ""
		if t, ok := titles[c.PostId]; ok {
			title = t
		} else {
			title = service.GetPostTitle(c.PostId)
			titles[c.PostId] = title
		}
		content := c.Source
		if content == "" {
			content = c.Content
		}
		cmts = append(cmts, &Comment{
			Comment:   c,
			PostTitle: title,
			Content:   content,
		})
	}
	return cmts
}

func (b *Blog) listPostComments(c *gin.Context) {
	userCtx := b.auth.AuthCookie(c).Context(nil)

	name := utils.MustToInt64(c.Param("name"))
	limit := utils.MustToInt64(c.DefaultQuery("limit", "10"))
	offset := utils.MustToInt64(c.DefaultQuery("offset", "0"))

	comments, err := b.service.ListComments(userCtx, &protocols.ListCommentsRequest{
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

func (b *Blog) createPostComment(c *gin.Context) {
	comment := &protocols.Comment{
		PostId:     utils.MustToInt64(c.Param("name")),
		Parent:     utils.MustToInt64(c.DefaultPostForm("parent", "0")),
		Author:     c.DefaultPostForm("author", ""),
		Email:      c.DefaultPostForm("email", ""),
		Url:        c.DefaultPostForm("url", ""),
		Ip:         c.ClientIP(),
		Date:       datetime.ProtoLocal(),
		SourceType: c.PostForm("source_type"),
		Source:     c.PostForm("source"),
	}
	user := b.auth.AuthCookie(c)
	comment = b.service.CreateComment(user.Context(nil), comment)
	c.JSON(200, comment)
}
