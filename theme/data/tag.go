package data

import (
	"context"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/service"
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
		User:    auth.Context(ctx).User,
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
