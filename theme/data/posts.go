package data

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// PostsData ...
type PostsData struct {
	Posts        []*Post
	PostCount    int64
	PageCount    int64
	CommentCount int64
	ViewCount    int64
}

// NewDataForPosts ...
func NewDataForPosts(cfg *config.Config, user *auth.User, service *service.Service, r *http.Request) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}

	postsData := &PostsData{
		PostCount:    service.GetDefaultIntegerOption("post_count", 0),
		PageCount:    service.GetDefaultIntegerOption("page_count", 0),
		CommentCount: service.GetDefaultIntegerOption("comment_count", 0),
	}

	s := r.URL.Query().Get(`sort`)
	if s == `` {
		s = `date.desc`
	}

	sort := strings.SplitN(s, ".", 2)
	if len(sort) != 2 {
		sort = []string{"date", "desc"}
	}
	if !utils.StrInSlice([]string{"id", "title", "date", "page_view", "comments"}, sort[0]) {
		sort[0] = "date"
	}
	if !utils.StrInSlice([]string{"asc", "desc"}, sort[1]) {
		sort[1] = "desc"
	}

	posts := service.MustListPosts(user.Context(context.TODO()),
		&protocols.ListPostsRequest{
			Fields:  "id,title,date,page_view,comments,status",
			OrderBy: fmt.Sprintf(`%s %s`, sort[0], sort[1]),
		})

	for _, p := range posts {
		postsData.ViewCount += int64(p.PageView)
	}

	for _, p := range posts {
		pp := newPost(p)
		pp.link = service.GetLink(pp.ID)
		postsData.Posts = append(postsData.Posts, pp)
	}
	d.Posts = postsData
	return d
}
