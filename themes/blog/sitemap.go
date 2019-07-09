package blog

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
func (b *Blog) GetSitemap(c *gin.Context) {
	user := b.auth.AuthCookie(c)

	if ifModified := c.GetHeader("If-Modified-Since"); ifModified != "" {
		if modified := b.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
			if ifModified == datetime.My2Gmt(modified) {
				c.Status(http.StatusNotModified)
				return
			}
		}
	}

	rawPosts := b.service.MustListPosts(user.Context(nil),
		&protocols.ListPostsRequest{
			Fields:  "id",
			OrderBy: "date DESC",
		})

	sitemapPosts := make([]*PostForSitemap, 0, len(rawPosts))
	for _, post := range rawPosts {
		sitemapPosts = append(sitemapPosts, &PostForSitemap{
			Post: post,
			Link: fmt.Sprintf("%s/%d/", b.service.HomeURL(), post.ID),
		})
	}

	data := SitemapData{
		Posts: sitemapPosts,
	}

	buf := bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.render(buf, "sitemap", data)
	str := buf.String()
	c.Header("Content-Type", "application/xml")
	if modified := b.service.GetDefaultStringOption("last_post_time", ""); modified != "" {
		c.Header("Last-Modified", datetime.My2Gmt(modified))
	}
	c.String(200, "%s", str)
}
