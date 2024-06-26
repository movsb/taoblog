package data

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

// TagsData ...
type TagsData struct {
	Tags []*models.TagWithCount
}

// NewDataForTags ...
func NewDataForTags(ctx context.Context, cfg *config.Config, user *auth.User, service proto.TaoBlogServer, impl service.ToBeImplementedByRpc) *Data {
	d := &Data{
		ctx:    ctx,
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}
	tags := impl.ListTagsWithCount()
	td := &TagsData{
		Tags: tags,
	}
	d.Tags = td
	return d
}
