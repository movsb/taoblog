package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/movsb/taoblog/modules/taorm"
	"github.com/movsb/taoblog/service/models"
)

func (s *ImplServer) GetOption(name string) *models.Option {
	query := `SELECT * FROM options WHERE name = ?`
	var option models.Option
	taorm.MustQueryRows(&option, s.db, query, name)
	return &option
}

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

func (s *ImplServer) HaveOption(name string) (have bool) {
	defer func() {
		if e := recover(); e != nil {
			have = false
		}
	}()
	s.GetStringOption(name)
	return true
}

func (s *ImplServer) SetOption(name string, value interface{}) {
	if s.HaveOption(name) {
		option := s.GetOption(name)
		stmt := s.tdb.Model(option, "options")
		stmt.UpdateField("value", value)
	} else {
		option := models.Option{
			Name:  name,
			Value: fmt.Sprint(value),
		}
		s.tdb.Model(&option, "options").Create()
	}
}
