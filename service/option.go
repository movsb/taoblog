package service

import (
	"fmt"
	"strconv"

	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm/taorm"
)

func optionCacheKey(name string) string {
	return "option:" + name
}

func (s *Service) GetStringOption(name string) string {
	val, err := s.cache.Get(optionCacheKey(name), func(key string) (interface{}, error) {
		var option models.Option
		s.tdb.Model(models.Option{}).Where("name=?", name).MustFind(&option)
		return option.Value, nil
	})
	if err != nil {
		panic(err)
	}
	return val.(string)
}

func (s *Service) GetDefaultStringOption(name string, def string) string {
	val, err := s.cache.Get(optionCacheKey(name), func(key string) (interface{}, error) {
		var option models.Option
		err := s.tdb.Where("name=?", name).Find(&option)
		if err != nil {
			if taorm.IsNotFoundError(err) {
				return def, nil
			}
			return nil, err
		}
		return option.Value, nil
	})
	if err != nil {
		panic(err)
	}
	return val.(string)
}

func (s *Service) GetDefaultIntegerOption(name string, def int64) (value int64) {
	val, err := s.cache.Get(optionCacheKey(name), func(key string) (interface{}, error) {
		var option models.Option
		err := s.tdb.Model(models.Option{}).Where("name=?", name).Find(&option)
		if err != nil {
			if taorm.IsNotFoundError(err) {
				return def, nil
			}
			return nil, err
		}
		return strconv.ParseInt(option.Value, 10, 64)
	})
	if err != nil {
		panic(err)
	}
	return val.(int64)
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
}

func (s *Service) GetOption(name string) (*models.Option, error) {
	var option models.Option
	err := s.tdb.Model(models.Option{}).Where("name=?", name).Find(&option)
	return &option, err
}
