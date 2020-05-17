package blog

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

// GetRss ...
func (b *Blog) GetRss(c *gin.Context) {
	user := b.auth.AuthCookie(c)

	if ifModified := c.GetHeader("If-Modified-Since"); ifModified != "" {
		if modified := b.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
			if ifModified == datetime.My2Gmt(modified) {
				c.Status(http.StatusNotModified)
				return
			}
		}
	}

	posts := b.service.GetLatestPosts(user.Context(nil), "id,title,date,content", 10)

	data := RssData{
		BlogName:    b.service.Name(),
		Home:        b.service.HomeURL(),
		Description: b.service.GetDefaultStringOption("desc", ""),
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
	b.render(buf, "rss", data)
	str := buf.String()
	c.Header("Content-Type", "application/xml")
	if modified := b.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
		c.Header("Last-Modified", datetime.My2Gmt(modified))
	}
	c.String(200, "%s", str)
}
