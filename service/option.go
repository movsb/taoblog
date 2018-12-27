package service

import (
	"database/sql"
	"strconv"

	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/service/models"
)

func (s *ImplServer) GetStringOption(name string) string {
	query := `SELECT * FROM options WHERE name = ?`
	var option models.Option
	taorm.MustQueryRows(&option, s.db, query, name)
	return option.Value
}

func (s *ImplServer) GetDefaultStringOption(name string, def string) string {
	query := `SELECT * FROM options WHERE name = ?`
	var option models.Option
	switch err := taorm.QueryRows(&option, s.db, query, name); err {
	case nil:
		return option.Value
	case sql.ErrNoRows:
		return def
	default:
		panic(err)
	}
}

func (s *ImplServer) GetIntegerOption(name string) (value int64) {
	query := `SELECT * FROM options WHERE name = ?`
	var option models.Option
	taorm.MustQueryRows(&option, s.db, query, name)
	if n, err := strconv.ParseInt(option.Value, 10, 64); err != nil {
		panic(err)
		value = n
	}
	return
}

func (s *ImplServer) GetDefaultIntegerOption(name string, def int64) (value int64) {
	query := `SELECT * FROM options WHERE name = ?`
	var option models.Option
	switch err := taorm.QueryRows(&option, s.db, query, name); err {
	case nil:
		n, err := strconv.ParseInt(option.Value, 10, 64)
		if err != nil {
			panic(err)
		}
		value = n
		return
	case sql.ErrNoRows:
		return def
	default:
		panic(err)
	}
}
