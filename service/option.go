package service

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v2"
)

func optionCacheKey(name string) string {
	return "option:" + name
}

func (s *Service) GetStringOption(name string) (string, error) {
	val, err, _ := s.cache.GetOrLoad(context.TODO(), optionCacheKey(name),
		func(ctx context.Context, _ string) (any, time.Duration, error) {
			val, err := s._getOption(name)
			return val, time.Hour, err
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
			return n, time.Hour, err
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

func (s *Service) SetOption(name string, value any) {
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
		stmt.MustUpdateMap(map[string]any{
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

func (s *Service) GetConfig(ctx context.Context, req *proto.GetConfigRequest) (*proto.GetConfigResponse, error) {
	s.MustBeAdmin(ctx)

	u := config.NewUpdater(s.cfg)
	p := u.Find(req.Path)

	y, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}

	return &proto.GetConfigResponse{
		Yaml: string(y),
	}, nil
}

func (s *Service) SetConfig(ctx context.Context, req *proto.SetConfigRequest) (*proto.SetConfigResponse, error) {
	s.MustBeAdmin(ctx)

	u := config.NewUpdater(s.cfg)
	u.MustApply(req.Path, req.Yaml, func(path, value string) {
		s.SetOption(path, value)
		log.Println(`保存：`, path, value)
	})
	return &proto.SetConfigResponse{}, nil
}

type _PluginStorage struct {
	ss *Service
	ns string
}

func (s *_PluginStorage) Set(key string, value string) error {
	name := s.ns + `:` + key
	s.ss.SetOption(name, value)
	return nil
}

func (s *_PluginStorage) Get(key string) (string, error) {
	name := s.ns + `:` + key
	return s.ss.GetStringOption(name)
}

func (s *Service) GetPluginStorage(name string) utils.PluginStorage {
	return &_PluginStorage{
		ss: s,
		ns: name,
	}
}

func (s *Service) Restart(ctx context.Context, req *proto.RestartRequest) (*proto.RestartResponse, error) {
	s.MustBeAdmin(ctx)

	if s.cancel == nil {
		return nil, status.Error(codes.Unimplemented, `服务器不支持此操作。`)
	}

	s.maintenance.Enter(req.Reason, time.Second*10)

	// 延迟重启可以基本保证 grpc 响应发送完成，不至于使客户端报错。
	time.AfterFunc(time.Second*3, s.cancel)

	return &proto.RestartResponse{}, nil
}

func (s *Service) ScheduleUpdate(ctx context.Context, req *proto.ScheduleUpdateRequest) (*proto.ScheduleUpdateResponse, error) {
	s.MustBeAdmin(ctx)

	if s.cancel == nil {
		return nil, status.Error(codes.Unimplemented, `服务器不支持此操作。`)
	}

	s.scheduledUpdate.Store(true)
	log.Println(`已设置计划更新标识。`)

	// 如果一分钟内没有更新，自动重启。
	// 因为没有解决如何取消这个状态的函数。
	time.AfterFunc(time.Minute, s.cancel)

	return &proto.ScheduleUpdateResponse{}, nil
}
