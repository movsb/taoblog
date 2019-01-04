package front

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/service"
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

		},
	}
	pageData := &PagePostsData{}
	pageData.Posts = f.server.MustListPosts(&service.ListPostsRequest{
		Fields:  "id,title,date,page_view,comments",
		OrderBy: "date DESC",
	})

	f.render(c.Writer, "header", header)
	f.render(c.Writer, "page_posts", pageData)
	f.render(c.Writer, "footer", footer)
}
