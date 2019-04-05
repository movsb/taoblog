package blog

import (
	"strings"

	"github.com/movsb/taoblog/modules/datetime"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

// Comment ...
type Comment struct {
	*models.Comment
	PostTitle string
	IsAdmin   bool
}

func newComments(comments []*models.Comment, service *service.Service) []*Comment {
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

// AjaxComment ...
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

// NewAjaxComment ...
func NewAjaxComment(c *models.Comment, logged bool, adminEmail string) *AjaxComment {
	a := AjaxComment{
		ID:       c.ID,
		Parent:   c.Parent,
		Ancestor: c.Ancestor,
		PostID:   c.PostID,
		Author:   c.Author,
		URL:      c.URL,
		Date:     c.Date,
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

func (b *Blog) listPostComments(c *gin.Context) {
	userCtx := b.auth.AuthCookie(c).Context(nil)
	name := utils.MustToInt64(c.Param("name"))
	limit := utils.MustToInt64(c.DefaultQuery("limit", "10"))
	offset := utils.MustToInt64(c.DefaultQuery("offset", "0"))
	parents := b.service.ListComments(userCtx, &protocols.ListCommentsRequest{
		PostID:   name,
		Ancestor: 0,
		Limit:    limit,
		Offset:   offset,
		OrderBy:  "id DESC",
	})
	childrenMap := make(map[int64][]*models.Comment)
	for _, parent := range parents {
		childrenMap[parent.ID] = b.service.ListComments(userCtx, &protocols.ListCommentsRequest{
			PostID:   name,
			Ancestor: parent.ID,
			OrderBy:  "id ASC",
		})
	}

	user := b.auth.AuthCookie(c)
	adminEmail := b.service.GetStringOption("email")

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

func (b *Blog) createPostComment(c *gin.Context) {
	cmt := models.Comment{
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
	b.service.CreateComment(user.Context(nil), &cmt)
	adminEmail := b.service.GetDefaultStringOption("email", "")
	c.JSON(200, NewAjaxComment(&cmt, !user.IsGuest(), adminEmail))
}
