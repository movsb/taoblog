package data

import (
	"github.com/movsb/taoblog/config"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/service"
)

// SearchData ...
type SearchData struct {
	EngineID string
}

// NewDataForSearch ...
func NewDataForSearch(cfg *config.Config, user *auth.User, service *service.Service) *Data {
	d := &Data{
		Config: cfg,
		User:   user,
		Meta:   &MetaData{},
	}

	d.Search = &SearchData{
		EngineID: cfg.Site.Search.EngineID,
	}

	return d
}
