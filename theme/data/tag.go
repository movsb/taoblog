package data

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	proto "github.com/movsb/taoblog/protocols"
)

// TagData ...
type TagData struct {
	Names []string
	Posts []*Post
}

// NewDataForTag ...
func NewDataForTag(ctx context.Context, cfg *config.Config, service proto.TaoBlogServer, tags []string) *Data {
	d := &Data{
		ctx:    ctx,
		Config: cfg,
		User:   auth.Context(ctx).User,
		Meta:   &MetaData{},
	}
	td := &TagData{
		Names: tags,
	}
	posts, err := service.GetPostsByTags(ctx,
		&proto.GetPostsByTagsRequest{
			Tags:     tags,
			WithLink: proto.LinkKind_LinkKindRooted,
		},
	)
	if err != nil {
		panic(err)
	}
	for _, p := range posts.Posts {
		pp := newPost(p)
		td.Posts = append(td.Posts, pp)
	}
	d.Tag = td
	return d
}
