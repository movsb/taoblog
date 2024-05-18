package data

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service"
)

// TagData ...
type TagData struct {
	Names []string
	Posts []*Post
}

// NewDataForTag ...
func NewDataForTag(ctx context.Context, cfg *config.Config, service protocols.TaoBlogServer, impl service.ToBeImplementedByRpc, tags []string) *Data {
	d := &Data{
		Config: cfg,
		User:   auth.Context(ctx).User,
		Meta:   &MetaData{},
	}
	td := &TagData{
		Names: tags,
	}
	posts, err := service.GetPostsByTags(ctx, &protocols.GetPostsByTagsRequest{Tags: tags})
	if err != nil {
		panic(err)
	}
	for _, p := range posts.Posts {
		pp := newPost(p)
		pp.link = impl.GetLink(p.Id)
		td.Posts = append(td.Posts, pp)
	}
	d.Tag = td
	return d
}
