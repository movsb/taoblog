package data

import (
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/models"
)

// TagsData ...
type TagsData struct {
	Tags []*models.TagWithCount
}

// NewDataForTags ...
func NewDataForTags(cfg *config.Config, user *auth.User, service *service.Service) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}
	tags := service.ListTagsWithCount()
	td := &TagsData{
		Tags: tags,
	}
	d.Tags = td
	return d
}
