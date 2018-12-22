package service

import (
	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
)

func (s *ImplServer) GetOption(in *protocols.GetOptionRequest) *protocols.Option {
	query := `SELECT * FROM options WHERE name = ?`
	var option models.Option
	taorm.MustQueryRows(&option, s.db, query, in.Name)
	return option.Serialize()
}

func (s *ImplServer) ListOptions(in *protocols.ListOptionsRequest) *protocols.ListOptionsResponse {
	query := `SELECT * FROM options`
	var options models.Options
	taorm.MustQueryRows(&options, s.db, query)
	return &protocols.ListOptionsResponse{
		Options: options.Serialize(),
	}
}
