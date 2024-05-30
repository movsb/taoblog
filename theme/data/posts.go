package data

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	proto "github.com/movsb/taoblog/protocols"
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
func NewDataForPosts(ctx context.Context, cfg *config.Config, service proto.TaoBlogServer, impl service.ToBeImplementedByRpc, r *http.Request) *Data {
	d := &Data{
		ctx:    ctx,
		Config: cfg,
		User:   auth.Context(ctx).User,
		Meta:   &MetaData{},
	}

	postsData := &PostsData{
		PostCount:    impl.GetDefaultIntegerOption("post_count", 0),
		PageCount:    impl.GetDefaultIntegerOption("page_count", 0),
		CommentCount: impl.GetDefaultIntegerOption("comment_count", 0),
	}

	s := r.URL.Query().Get(`sort`)
	if s == `` {
		s = `date.desc`
	}

	sort := strings.SplitN(s, ".", 2)
	if len(sort) != 2 {
		sort = []string{"date", "desc"}
	}
	if !slices.Contains([]string{"id", "title", "date", "page_view", "comments"}, sort[0]) {
		sort[0] = "date"
	}
	if !slices.Contains([]string{"asc", "desc"}, sort[1]) {
		sort[1] = "desc"
	}

	posts, err := service.ListPosts(ctx,
		&proto.ListPostsRequest{
			OrderBy:  fmt.Sprintf(`%s %s`, sort[0], sort[1]),
			Kinds:    []string{`post`},
			WithLink: proto.LinkKind_LinkKindRooted,
		},
	)
	if err != nil {
		panic(err)
	}

	for _, p := range posts.Posts {
		postsData.ViewCount += int64(p.PageView)
	}

	for _, p := range posts.Posts {
		pp := newPost(p)
		postsData.Posts = append(postsData.Posts, pp)
	}
	d.Posts = postsData
	return d
}
