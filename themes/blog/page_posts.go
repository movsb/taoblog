package blog

import (
	"fmt"
	"strings"

	"github.com/movsb/taoblog/modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

// PagePostsData ...
type PagePostsData struct {
	Posts []*protocols.Post
}

func (b *Blog) getPagePosts(c *gin.Context) {
	user := b.auth.AuthCookie(c)
	header := &ThemeHeaderData{
		Title: "全部文章",
		Header: func() {
			b.render(c.Writer, "page_posts_header", nil)
		},
	}
	footer := &ThemeFooterData{
		Footer: func() {
			b.render(c.Writer, "page_posts_footer", nil)
		},
	}
	pageData := &PagePostsData{}

	sort := strings.SplitN(c.DefaultQuery("sort", "date.desc"), ".", 2)
	if len(sort) != 2 {
		sort = []string{"date", "desc"}
	}
	if !utils.StrInSlice([]string{"id", "title", "date", "page_view", "comments"}, sort[0]) {
		sort[0] = "date"
	}
	if !utils.StrInSlice([]string{"asc", "desc"}, sort[1]) {
		sort[1] = "desc"
	}

	pageData.Posts = b.service.MustListPosts(user.Context(nil),
		&protocols.ListPostsRequest{
			Fields:  "id,title,date,page_view,comments",
			OrderBy: fmt.Sprintf(`%s %s`, sort[0], sort[1]),
		})

	b.render(c.Writer, "header", header)
	b.render(c.Writer, "page_posts", pageData)
	b.render(c.Writer, "footer", footer)
}
