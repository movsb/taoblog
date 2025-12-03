package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

func optionCacheKey(name string) string {
	return "option:" + name
}

func (s *Service) Options() utils.PluginStorage {
	if s.options == nil {
		panic(`未实现`)
	}
	return s.options
}

func (s *Service) getOption(name string) (string, error) {
	val, err, _ := s.cache.GetOrLoad(context.Background(), optionCacheKey(name),
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

func (s *Service) setOption(name string, value string) error {
	if s._haveOption(name) {
		stmt := s.tdb.From(models.Option{}).Where("name = ?", name)
		stmt.MustUpdateMap(map[string]any{
			"value": value,
		})
	} else {
		option := models.Option{
			Name:  name,
			Value: value,
		}
		s.tdb.Model(&option).MustCreate()
	}
	s.cache.Set(optionCacheKey(name), value, time.Hour)
	return nil
}

// https://blog.twofei.com/869/
var _sqlEscapeReplacer = strings.NewReplacer(`%`, `\%`, `_`, `\_`, `\`, `\\`)

// prefix：要么空，要么带冒号。
func (s *Service) rangeOptions(prefix string, iter func(key string)) error {
	var options []*models.Option
	var like string
	if prefix == `` {
		like = `%%`
	} else {
		like = _sqlEscapeReplacer.Replace(prefix)
		like = fmt.Sprintf(`%s%%`, like)
	}
	s.tdb.Select(`name`).Where(`name like ?`, like).MustFind(&options)
	for _, opt := range options {
		iter(opt.Name)
	}
	return nil
}

func (s *Service) GetDefaultIntegerOption(name string, def int64) int64 {
	return utils.Must1(s.options.GetIntegerDefault(name, def))
}

func (s *Service) GetConfig(ctx context.Context, req *proto.GetConfigRequest) (*proto.GetConfigResponse, error) {
	user.MustBeAdmin(ctx)

	if strings.HasPrefix(req.Path, runtime_config.Prefix) || req.Path+`.` == runtime_config.Prefix {
		return s.runtime.GetConfig(ctx, req)
	}

	if strings.HasPrefix(req.Path, `/`) {
		return s.userRoots.GetConfig(ctx, req)
	}

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
	user.MustBeAdmin(ctx)

	if strings.HasPrefix(req.Path, runtime_config.Prefix) {
		return s.runtime.SetConfig(ctx, req)
	}

	if strings.HasPrefix(req.Path, `/`) {
		return s.userRoots.SetConfig(ctx, req)
	}

	u := config.NewUpdater(s.cfg)
	u.MustApply(req.Path, req.Yaml, func(path, value string) {
		utils.Must(s.options.SetString(path, value))
		log.Println(`保存：`, path, value)
	})
	return &proto.SetConfigResponse{}, nil
}

type _PluginStorage struct {
	ss     *Service
	prefix string
}

func (s *_PluginStorage) SetString(key string, value string) error {
	return s.ss.setOption(s.prefix+key, value)
}

func (s *_PluginStorage) GetString(key string) (string, error) {
	return s.ss.getOption(s.prefix + key)
}

func (s *_PluginStorage) GetStringDefault(key string, def string) (string, error) {
	value, err := s.GetString(key)
	if err == nil {
		return value, nil
	}
	if taorm.IsNotFoundError(err) {
		return def, nil
	}
	return ``, err
}

func (s *_PluginStorage) SetInteger(key string, value int64) error {
	return s.SetString(key, fmt.Sprint(value))
}

func (s *_PluginStorage) GetInteger(key string) (int64, error) {
	str, err := s.GetString(key)
	if err == nil {
		return strconv.ParseInt(str, 10, 64)
	}
	return 0, err
}

func (s *_PluginStorage) GetIntegerDefault(key string, def int64) (int64, error) {
	value, err := s.GetInteger(key)
	if err == nil {
		return value, nil
	}
	if taorm.IsNotFoundError(err) {
		return def, nil
	}
	return 0, err
}

func (s *_PluginStorage) Range(iter func(key string)) {
	s.ss.rangeOptions(s.prefix, func(key string) {
		iter(strings.TrimPrefix(key, s.prefix))
	})
}

func (s *Service) GetPluginStorage(name string) utils.PluginStorage {
	prefix := ``
	if name != `` {
		prefix = name + `:`
	}
	return &_PluginStorage{
		ss:     s,
		prefix: prefix,
	}
}

func (s *Service) Restart(ctx context.Context, req *proto.RestartRequest) (*proto.RestartResponse, error) {
	user.MustBeAdmin(ctx)

	if s.cancel == nil {
		return nil, status.Error(codes.Unimplemented, `服务器不支持此操作。`)
	}

	s.maintenance.Enter(req.Reason, time.Second*10)

	// 延迟重启可以基本保证 grpc 响应发送完成，不至于使客户端报错。
	time.AfterFunc(time.Second*3, s.cancel)

	return &proto.RestartResponse{}, nil
}

func (s *Service) ScheduleUpdate(ctx context.Context, req *proto.ScheduleUpdateRequest) (*proto.ScheduleUpdateResponse, error) {
	user.MustBeAdmin(ctx)

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

func (s *Service) GetSiteConfig(ctx context.Context, in *proto.GetSiteConfigRequest) (*proto.GetSiteConfigResponse, error) {
	user.MustBeAdmin(ctx)

	c := proto.SiteConfig{
		Name:        s.Config().Site.GetName(),
		Description: s.Config().Site.GetDescription(),

		// TODO 可能是覆盖值，不允许前端修改。
		Home: s.Config().Site.GetHome(),

		Icon: utils.CreateDataURL(s.Favicon().Data).String(),
	}
	return &proto.GetSiteConfigResponse{
		Config: &c,
	}, nil
}

// 网站图标默认尺寸。
const IconSize = 100

func (s *Service) UpdateSiteConfig(ctx context.Context, in *proto.UpdateSiteConfigRequest) (*proto.UpdateSiteConfigResponse, error) {
	user.MustBeAdmin(ctx)

	u := config.NewUpdater(s.cfg)
	store := s.GetPluginStorage(`config`)
	save := func(path string, value string) {
		utils.Must(store.SetString(path, value))
	}

	if name := in.Config.Name; strings.TrimSpace(name) == `` || len(name) > 100 {
		return nil, status.Errorf(codes.InvalidArgument, `网站名称不能为空，且不能超过 100 字符。`)
	} else {
		u.MustApply(`site.name`, in.Config.Name, save)
	}

	if desc := in.Config.Description; len(desc) > 500 {
		return nil, status.Errorf(codes.InvalidArgument, `网站描述不能超过 500 字符。`)
	} else {
		u.MustApply(`site.description`, in.Config.Description, save)
	}

	if homeURL, err := url.Parse(in.Config.Home); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, `网站首页 URL 错误：%v`, err)
	} else {
		if homeURL.Scheme != `http` && homeURL.Scheme != `https` {
			return nil, status.Errorf(codes.InvalidArgument, `网站首页 URL 必须是 “http” 或 “https” 协议。`)
		}
		if homeURL.Path != `` && homeURL.Path != `/` {
			return nil, status.Errorf(codes.InvalidArgument, `网站首页 URL 不能带路径部分。`)
		}
		u.MustApply(`site.home`, in.Config.Home, save)
	}

	rspConfig := proto.UpdateSiteConfigResponse{
		Config: &proto.SiteConfig{
			Name:        in.Config.Name,
			Description: in.Config.Description,
			Home:        in.Config.Home,
		},
	}

	if in.UpdateIcon {
		d, err := utils.ParseDataURL(in.Config.Icon)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, `网站图标错误：%v`, err)
		}
		if !strings.HasPrefix(d.Type, `image/`) {
			return nil, status.Errorf(codes.InvalidArgument, `网站图标必须是图片格式。`)
		}
		if len(d.Data) > 100<<10 {
			d, err = utils.ResizeImage(d.Type, bytes.NewReader(d.Data), IconSize, IconSize)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, `网站图标处理失败：%v`, err)
			}
		}

		s.options.SetString(`favicon`, d.String())
		s.favicon.SetData(time.Now(), d)
		rspConfig.Config.Icon = d.String()
	}

	return &rspConfig, nil
}

var jsonPB = &runtime.JSONPb{
	MarshalOptions: protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	},
}

func (s *Service) GetUserSettings(ctx context.Context, in *proto.GetUserSettingsRequest) (_ *proto.Settings, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.MustNotBeGuest(ctx)

	ss := utils.Must1(s.options.GetStringDefault(fmt.Sprintf(`user_settings:%d`, ac.User.ID), `{}`))

	var out proto.Settings
	utils.Must(jsonPB.Unmarshal([]byte(ss), &out))

	{
		u := utils.Must1(url.Parse(s.getHome())).JoinPath(`/v3/calendars`)
		js := utils.Must1(json.Marshal(CalendarData{
			UserID: int(ac.User.ID),
		}))
		q := url.Values{}

		// 写固定 nonce 加密的数据，防止总是变化。
		nonce := [12]byte{}
		binary.LittleEndian.PutUint32(nonce[0:], uint32(ac.User.ID))
		binary.LittleEndian.PutUint32(nonce[4:], uint32(ac.User.ID))
		binary.LittleEndian.PutUint32(nonce[8:], uint32(ac.User.ID))
		d := utils.Must1(s.aesGCM.Encrypt(js, nonce[:]))
		encoded := base64.RawURLEncoding.EncodeToString(d)
		q.Set(`data`, encoded)

		u.RawQuery = q.Encode()
		out.CalendarUrl = u.String()
	}

	return &out, nil
}

func (s *Service) SetUserSettings(ctx context.Context, in *proto.SetUserSettingsRequest) (_ *proto.Settings, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.MustNotBeGuest(ctx)

	old := utils.Must1(s.GetUserSettings(ctx, &proto.GetUserSettingsRequest{}))

	if in.UpdateGroupPostsByCategory {
		old.GroupPostsByCategory = in.Settings.GroupPostsByCategory
	}

	by := utils.Must1(jsonPB.Marshal(old))
	utils.Must(s.options.SetString(fmt.Sprintf(`user_settings:%d`, ac.User.ID), string(by)))

	return s.GetUserSettings(ctx, &proto.GetUserSettingsRequest{})
}
