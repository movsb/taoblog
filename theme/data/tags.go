package data

import (
	"context"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

// TagsData ...
type TagsData struct {
	Tags []*models.TagWithCount
}

// NewDataForTags ...
func NewDataForTags(ctx context.Context, cfg *config.Config, service proto.TaoBlogServer, impl service.ToBeImplementedByRpc) *Data {
	d := &Data{
		ctx:    ctx,
		Config: cfg,
		Meta:   &MetaData{},
	}
	tags := impl.ListTagsWithCount()
	td := &TagsData{
		Tags: tags,
	}
	d.Tags = td
	return d
}
