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
	td := &TagData{
		Names: tags,
	}
	posts := service.GetPostsByTags(tags).ToProtocols()
	for _, p := range posts {
		pp := newPost(p)
		pp.link = service.GetLink(p.Id)
		td.Posts = append(td.Posts, pp)

	}
	d.Tag = td
	return d
}
