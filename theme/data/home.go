package data

import (
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// HomeData ...
type HomeData struct {
	Posts []*Post

	PostCount    int64
	PageCount    int64
	CommentCount int64
}

// NewDataForHome ...
func NewDataForHome(cfg *config.Config, user *auth.User, service *service.Service) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}
	home := &HomeData{
		PostCount:    service.GetDefaultIntegerOption("post_count", 0),
		PageCount:    service.GetDefaultIntegerOption("page_count", 0),
		CommentCount: service.GetDefaultIntegerOption("comment_count", 0),
	}
	home.Posts = newPosts(service.MustListPosts(user.Context(nil),
		&protocols.ListPostsRequest{
			Fields:  "id,title,type,status",
			Limit:   20,
			OrderBy: "date DESC",
		}))
	d.Home = home
	d.svc = service
	return d
}
