package front

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/movsb/taoblog/protocols"

	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/modules/datetime"
)

// PostForSitemap ...
type PostForSitemap struct {
	*protocols.Post
	Link string
}

// SitemapData ...
type SitemapData struct {
	Posts []*PostForSitemap
}

// GetSitemap ...
func (f *Front) GetSitemap(c *gin.Context) {
	user := f.auth.AuthCookie(c)

	if ifModified := c.GetHeader("If-Modified-Since"); ifModified != "" {
		if modified := f.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
			if ifModified == datetime.Local2Gmt(modified) {
				c.Status(http.StatusNotModified)
				return
			}
		}
	}

	rawPosts := f.service.MustListPosts(user.Context(nil),
		&protocols.ListPostsRequest{
			Fields:  "id",
			OrderBy: "date DESC",
		})

	home := "https://" + f.service.GetDefaultStringOption("home", "taoblog.local")
	sitemapPosts := make([]*PostForSitemap, 0, len(rawPosts))
	for _, post := range rawPosts {
		sitemapPosts = append(sitemapPosts, &PostForSitemap{
			Post: post,
			Link: fmt.Sprintf("%s/%d/", home, post.ID),
		})
	}

	data := SitemapData{
		Posts: sitemapPosts,
	}

	buf := bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>`)
	f.render(buf, "sitemap", data)
	str := buf.String()
	c.Header("Content-Type", "application/xml")
	if modified := f.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
		c.Header("Last-Modified", datetime.Local2Gmt(modified))
	}
	c.String(200, "%s", str)
}
