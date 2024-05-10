package service

import (
	"context"
	"strconv"
	"time"

	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
)

func optionCacheKey(name string) string {
	return "option:" + name
}

func (s *Service) GetStringOption(name string) (string, error) {
	val, err, _ := s.cache.GetOrLoad(context.TODO(), optionCacheKey(name),
		func(ctx context.Context, _ string) (any, time.Duration, error) {
			val, err := s._getOption(name)
			return val, time.Minute * 10, err
		},
	)
	if err != nil {
		return ``, err
	}
	return val.(string), nil
}

func (s *Service) GetDefaultStringOption(name string, def string) string {
	val, err := s.GetStringOption(name)
	if err == nil {
		return val
	}
	if taorm.IsNotFoundError(err) {
		return def
	}
	panic(err)
}

func (s *Service) GetIntegerOption(name string) (int64, error) {
	val, err, _ := s.cache.GetOrLoad(context.TODO(), optionCacheKey(name),
		func(ctx context.Context, _ string) (any, time.Duration, error) {
			val, err := s._getOption(name)
			if err != nil {
				return nil, 0, err
			}
			n, err := strconv.ParseInt(val, 10, 64)
			return n, time.Minute * 10, err
		},
	)
	if err != nil {
		return 0, err
	}
	return val.(int64), nil
}

func (s *Service) GetDefaultIntegerOption(name string, def int64) int64 {
	val, err := s.GetIntegerOption(name)
	if err == nil {
		return val
	}
	if taorm.IsNotFoundError(err) {
		return def
	}
	panic(err)
}

func (s *Service) _getOption(name string) (string, error) {
	var option models.Option
	if err := s.tdb.Model(models.Option{}).Where("name=?", name).Find(&option); err != nil {
		return ``, err
	}
	return option.Value, nil
}

func (s *Service) _haveOption(name string) (have bool) {
	_, err := s._getOption(name)
	return err == nil
}

func (s *Service) SetOption(name string, value interface{}) {
	var toSave string
	switch v := value.(type) {
	case string:
		toSave = v
	case int:
		toSave = strconv.Itoa(v)
		value = int64(v)
	case int64:
		toSave = strconv.FormatInt(v, 10)
	default:
		panic("unsupported option type:" + name)
	}
	if s._haveOption(name) {
		stmt := s.tdb.From(models.Option{}).Where("name = ?", name)
		stmt.MustUpdateMap(map[string]interface{}{
			"value": toSave,
		})
	} else {
		option := models.Option{
			Name:  name,
			Value: toSave,
		}
		s.tdb.Model(&option).MustCreate()
	}
	s.cache.Set(optionCacheKey(name), value, time.Minute*10)
}
