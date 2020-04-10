package service

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm/taorm"
)

func optionCacheKey(name string) string {
	return "option:" + name
}

func (s *Service) GetStringOption(name string) string {
	if val, ok := s.cache.Get(optionCacheKey(name)); ok {
		return val.(string)
	}
	var option models.Option
	s.tdb.Model(models.Option{}).Where("name=?", name).MustFind(&option)
	s.cache.Set(optionCacheKey(name), option.Value)
	return option.Value
}

func (s *Service) GetDefaultStringOption(name string, def string) string {
	if val, ok := s.cache.Get(optionCacheKey(name)); ok {
		return val.(string)
	}
	var option models.Option
	err := s.tdb.Where("name=?", name).Find(&option)
	if err == nil {
		s.cache.Set(optionCacheKey(name), option.Value)
		return option.Value
	} else if taorm.IsNotFoundError(err) {
		s.cache.Set(optionCacheKey(name), def)
		return def
	}
	panic(err)
}

/*
func (s *Service) GetIntegerOption(name string) (value int64) {
	var option models.Option
	s.tdb.Model(models.Option{}, "options").Where("name=?", name).MustFind(&option)
	if n, err := strconv.ParseInt(option.Value, 10, 64); err != nil {
		panic(err)
		value = n
	}
	return
}
*/

func (s *Service) GetDefaultIntegerOption(name string, def int64) (value int64) {
	parse := func(s string) int64 {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			panic(err)
		}
		return n
	}
	if val, ok := s.cache.Get(optionCacheKey(name)); ok {
		return parse(val.(string))
	}
	var option models.Option
	err := s.tdb.Model(models.Option{}).Where("name=?", name).Find(&option)
	switch err {
	case nil:
		s.cache.Set(optionCacheKey(name), option.Value)
		return parse(option.Value)
	case sql.ErrNoRows:
		s.cache.Set(optionCacheKey(name), fmt.Sprint(def))
		return def
	default:
		panic(err)
	}
}

func (s *Service) HaveOption(name string) (have bool) {
	defer func() {
		if e := recover(); e != nil {
			have = false
		}
	}()
	s.GetStringOption(name)
	return true
}

func (s *Service) SetOption(name string, value interface{}) {
	if s.HaveOption(name) {
		stmt := s.tdb.From(models.Option{}).Where("name = ?", name)
		stmt.MustUpdateMap(map[string]interface{}{
			"value": value,
		})
	} else {
		option := models.Option{
			Name:  name,
			Value: fmt.Sprint(value),
		}
		s.tdb.Model(&option).MustCreate()
	}
	s.cache.Set(optionCacheKey(name), fmt.Sprint(value))

	// TODO use hook
	if name == "login" {
		s.auth.SetLogin(fmt.Sprint(value))
	}
}

func (s *Service) GetOption(name string) (*models.Option, error) {
	var option models.Option
	err := s.tdb.Model(models.Option{}).Where("name=?", name).Find(&option)
	return &option, err
}
