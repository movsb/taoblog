package front

import (
	"html/template"
	"strings"

	"github.com/movsb/taoblog/modules/datetime"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

type Comment struct {
	*models.Comment
	server *service.ImplServer
}

func newComment(comment *models.Comment, server *service.ImplServer) *Comment {
	return &Comment{
		Comment: comment,
		server:  server,
	}
}

func newComments(comments []*models.Comment, server *service.ImplServer) []*Comment {
	cmts := []*Comment{}
	for _, c := range comments {
		cmts = append(cmts, newComment(c, server))
	}
	return cmts
}

func (c *Comment) AuthorString() string {
	mail := c.server.GetDefaultStringOption("email", "")
	if mail == c.Email {
		return "博主"
	}
	return c.Author
}

func (c *Comment) PostTitle() template.HTML {
	return template.HTML(c.server.GetPostTitle(c.PostID))
}

type AjaxComment struct {
	// From Comment
	ID       int64  `json:"id"`
	Parent   int64  `json:"parent"`
	Ancestor int64  `json:"ancestor"`
	PostID   int64  `json:"post_id"`
	Author   string `json:"author"`
	Email    string `json:"email,omitempty"`
	URL      string `json:"url"`
	IP       string `json:"ip,omitempty"`
	Date     string `json:"date"`
	Content  string `json:"content"`

	// Owned
	Children []*AjaxComment `json:"children"`
	Avatar   string         `json:"avatar"`
	IsAdmin  bool           `json:"is_admin"`
}

func NewAjaxComment(c *models.Comment, logged bool, adminEmail string) *AjaxComment {
	a := AjaxComment{
		ID:       c.ID,
		Parent:   c.Parent,
		Ancestor: c.Ancestor,
		PostID:   c.PostID,
		Author:   c.Author,
		URL:      c.URL,
		Date:     datetime.My2Local(c.Date),
		Content:  c.Content,
	}

	a.Avatar = utils.Md5Str(c.Email)
	a.IsAdmin = strings.EqualFold(c.Email, adminEmail)

	if logged {
		a.Email = c.Email
		a.IP = c.IP
	}

	return &a
}

func (f *Front) listPostComments(c *gin.Context) {
	name := utils.MustToInt64(c.Param("name"))
	limit := utils.MustToInt64(c.DefaultQuery("limit", "10"))
	offset := utils.MustToInt64(c.DefaultQuery("offset", "0"))
	parents := f.server.ListPostComments(&service.ListCommentsRequest{
		PostID:   name,
		Ancestor: 0,
		Limit:    limit,
		Offset:   offset,
		OrderBy:  "id DESC",
	})
	childrenMap := make(map[int64][]*models.Comment)
	for _, parent := range parents {
		childrenMap[parent.ID] = f.server.ListPostComments(&service.ListCommentsRequest{
			PostID:   name,
			Ancestor: parent.ID,
			OrderBy:  "id ASC",
		})
	}

	user := f.auth.AuthCookie(c)
	adminEmail := f.server.GetStringOption("email")

	outParents := make([]*AjaxComment, 0, len(parents))
	for _, parent := range parents {
		outParent := NewAjaxComment(parent, !user.IsGuest(), adminEmail)
		outParents = append(outParents, outParent)
		outParent.Children = make([]*AjaxComment, 0, 0)
		for _, child := range childrenMap[parent.ID] {
			outChild := NewAjaxComment(child, !user.IsGuest(), adminEmail)
			outParent.Children = append(outParent.Children, outChild)
		}
	}
	c.JSON(200, outParents)
}

func (f *Front) createPostComment(c *gin.Context) {
	cmt := models.Comment{
		PostID:  utils.MustToInt64(c.Param("name")),
		Parent:  utils.MustToInt64(c.DefaultPostForm("parent", "0")),
		Author:  c.DefaultPostForm("author", ""),
		Email:   c.DefaultPostForm("email", ""),
		URL:     c.DefaultPostForm("url", ""),
		IP:      c.ClientIP(),
		Date:    datetime.MyLocal(), // TODO
		Content: c.DefaultPostForm("content", ""),
	}
	user := f.auth.AuthCookie(c)
	f.server.CreateComment(user.Context(nil), &cmt)
	adminEmail := f.server.GetDefaultStringOption("email", "")
	c.JSON(200, NewAjaxComment(&cmt, !user.IsGuest(), adminEmail))
}

func (f *Front) deletePostComment(c *gin.Context) {
	_ = utils.MustToInt64(c.Param("name"))
	commentName := utils.MustToInt64(c.Param("comment_name"))
	user := f.auth.AuthCookie(c)
	f.server.DeleteComment(user.Context(nil), commentName)
}
