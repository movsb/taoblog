package data

import (
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/service"
)

// TagsData ...
type TagsData struct {
	Names []string
	Posts []*Post
}

// NewDataForTags ...
func NewDataForTags(cfg *config.Config, user *auth.User, service *service.Service, tags []string) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}
	posts := newPosts(service.GetPostsByTags(tags).ToProtocols())
	td := &TagsData{
		Names: tags,
		Posts: posts,
	}
	d.Tags = td
	return d
}
