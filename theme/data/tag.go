package data

import (
	"context"

	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/micros/auth/user"
)

// TagData ...
type TagData struct {
	Names []string
	Posts []*Post
}

// NewDataForTag ...
func NewDataForTag(ctx context.Context, service service.ToBeImplementedByRpc, tags []string) *Data {
	d := &Data{
		Context: ctx,
		User:    user.Context(ctx).User,
	}
	td := &TagData{
		Names: tags,
	}
	posts, err := service.GetPostsByTags(ctx, tags)
	if err != nil {
		panic(err)
	}
	for _, p := range posts {
		pp := newPost(p)
		td.Posts = append(td.Posts, pp)
	}
	d.Data = td
	return d
}
