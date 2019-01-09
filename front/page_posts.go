package front

import (
	"fmt"
	"strings"

	"github.com/movsb/taoblog/modules/utils"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
)

type PagePostsData struct {
	Posts []*models.Post
}

func (f *Front) GetPagePosts(c *gin.Context) {
	header := &ThemeHeaderData{
		Title: "全部文章",
		Header: func() {
			f.render(c.Writer, "page_posts_header", nil)
		},
	}
	footer := &ThemeFooterData{
		Footer: func() {
			f.render(c.Writer, "page_posts_footer", nil)
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

	pageData.Posts = f.server.MustListPosts(&protocols.ListPostsRequest{
		Fields:  "id,title,date,page_view,comments",
		OrderBy: fmt.Sprintf(`%s %s`, sort[0], sort[1]),
	})

	f.render(c.Writer, "header", header)
	f.render(c.Writer, "page_posts", pageData)
	f.render(c.Writer, "footer", footer)
}
