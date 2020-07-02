package data

import (
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/service"
)

// TagsData ...
type TagsData struct {
	Name  string
	Posts []*Post
}

// NewDataForTags ...
func NewDataForTags(cfg *config.Config, user *auth.User, service *service.Service, tag string) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}
	posts := newPosts(service.GetPostsByTags(tag).ToProtocols())
	td := &TagsData{
		Name:  tag,
		Posts: posts,
	}
	d.Tags = td
	return d
}
