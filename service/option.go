package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/movsb/taoblog/service/models"
)

func (s *ImplServer) GetOption(name string) *models.Option {
	var option models.Option
	s.tdb.Model(models.Option{}, "options").Where("name=?", name).MustFind(&option)
	return &option
}

func (s *ImplServer) GetStringOption(name string) string {
	return s.GetOption(name).Value
}

func (s *ImplServer) GetDefaultStringOption(name string, def string) string {
	var option models.Option
	err := s.tdb.Model(models.Option{}, "options").Where("name=?", name).Find(&option)
	switch err {
	case nil:
		return option.Value
	case sql.ErrNoRows:
		return def
	default:
		panic(err)
	}
}

func (s *ImplServer) GetIntegerOption(name string) (value int64) {
	var option models.Option
	s.tdb.Model(models.Option{}, "options").Where("name=?", name).MustFind(&option)
	if n, err := strconv.ParseInt(option.Value, 10, 64); err != nil {
		panic(err)
		value = n
	}
	return
}

func (s *ImplServer) GetDefaultIntegerOption(name string, def int64) (value int64) {
	var option models.Option
	err := s.tdb.Model(models.Option{}, "options").Where("name=?", name).Find(&option)
	switch err {
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
		stmt.MustUpdateMap(map[string]interface{}{
			"value": value,
		})
	} else {
		option := models.Option{
			Name:  name,
			Value: fmt.Sprint(value),
		}
		s.tdb.Model(&option, "options").Create()
	}
}
