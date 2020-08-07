package data

import (
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/service"
)

// TagData ...
type TagData struct {
	Names []string
	Posts []*Post
}

// NewDataForTag ...
func NewDataForTag(cfg *config.Config, user *auth.User, service *service.Service, tags []string) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}
	posts := newPosts(service.GetPostsByTags(tags).ToProtocols())
	td := &TagData{
		Names: tags,
		Posts: posts,
	}
	d.Tag = td
	return d
}
