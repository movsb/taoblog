package front

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/movsb/taoblog/modules/datetime"
	"github.com/movsb/taoblog/protocols"

	"github.com/gin-gonic/gin"
)

// PostForRss for Rss Post.
type PostForRss struct {
	*protocols.Post
	Date    string
	Content template.HTML
	Link    string
}

// RssData for Rss template.
type RssData struct {
	BlogName    string
	Home        string
	Description string
	Posts       []*PostForRss
}

func (f *Front) GetRss(c *gin.Context) {
	if ifModified := c.GetHeader("If-Modified-Since"); ifModified != "" {
		if modified := f.server.GetDefaultStringOption("last_post_time", ""); modified != "" {
			if ifModified == datetime.Local2Gmt(modified) {
				c.Status(http.StatusNotModified)
				return
			}
		}
	}

	posts := f.server.GetLatestPosts("id,title,date,content", 10)

	data := RssData{
		BlogName:    f.server.GetDefaultStringOption("blog_name", "TaoBlog"),
		Home:        "https://" + f.server.GetDefaultStringOption("home", "taoblog.local"),
		Description: f.server.GetDefaultStringOption("desc", ""),
	}

	var rssPosts []*PostForRss
	for _, post := range posts {
		content := template.HTML("<![CDATA[" + strings.Replace(string(post.Content), "]]>", "]]]]><!CDATA[>", -1) + "]]>")
		link := fmt.Sprintf("%s/%d/", data.Home, post.ID)
		rssPosts = append(rssPosts, &PostForRss{
			Post:    post,
			Date:    datetime.My2Feed(post.Date),
			Content: content,
			Link:    link,
		})
	}

	data.Posts = rssPosts

	// html/template will break `<?xml` into `&lt;?xml`, we cannot use it here.
	// TODO use package encoding/xml instead.
	// See: https://github.com/golang/go/issues/3133
	buf := bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>`)
	f.render(buf, "rss", data)
	str := buf.String()
	c.Header("Content-Type", "application/xml")
	if modified := f.server.GetDefaultStringOption("last_post_time", ""); modified != "" {
		c.Header("Last-Modified", datetime.Local2Gmt(modified))
	}
	c.String(200, "%s", str)
}
