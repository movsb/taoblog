package data

import (
	"fmt"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// SitemapPostData ...
type SitemapPostData struct {
	*protocols.Post
	Link string
}

// SitemapData ...
type SitemapData struct {
	Posts []*SitemapPostData
}

// NewDataForSitemap ...
func NewDataForSitemap(cfg *config.Config, user *auth.User, service *service.Service) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}

	rawPosts := service.MustListPosts(user.Context(nil),
		&protocols.ListPostsRequest{
			Fields:  "id",
			OrderBy: "date DESC",
		})

	sitemapPosts := make([]*SitemapPostData, 0, len(rawPosts))
	for _, post := range rawPosts {
		sitemapPosts = append(sitemapPosts, &SitemapPostData{
			Post: post,
			Link: fmt.Sprintf("%s/%d/", service.HomeURL(), post.Id),
		})
	}

	sd := &SitemapData{
		Posts: sitemapPosts,
	}

	d.Sitemap = sd

	return d
}
