package runtime_config

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type runtimeKey struct{}

func Context(parent context.Context, r *Runtime) context.Context {
	return context.WithValue(parent, runtimeKey{}, r)
}

func FromContext(ctx context.Context) *Runtime {
	r, _ := ctx.Value(runtimeKey{}).(*Runtime)
	return r
}

type Runtime struct {
	l sync.Mutex
	m map[string]any
}

func NewRuntime() *Runtime {
	return &Runtime{
		m: map[string]any{},
	}
}

func (r *Runtime) Register(path string, obj any) {
	r.l.Lock()
	defer r.l.Unlock()

	if _, ok := r.m[path]; ok {
		panic(`运行时对象已经存在：` + path)
	}

	r.m[path] = obj
}

const Prefix = `runtime.`

func (r *Runtime) GetConfig(ctx context.Context, req *proto.GetConfigRequest) (*proto.GetConfigResponse, error) {
	r.l.Lock()
	defer r.l.Unlock()

	var p any

	if req.Path+`.` == Prefix {
		p = r.m
	} else {
		path := strings.TrimPrefix(req.Path, Prefix)
		module, path, _ := strings.Cut(path, `.`)

		obj := r.m[module]
		if obj == nil {
			return nil, errors.New(`无此运行时配置。`)
		}

		u := config.NewUpdater(obj)
		p = u.Find(path)
	}

	y, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}

	return &proto.GetConfigResponse{
		Yaml: string(y),
	}, nil
}

func (r *Runtime) SetConfig(ctx context.Context, req *proto.SetConfigRequest) (*proto.SetConfigResponse, error) {
	r.l.Lock()
	defer r.l.Unlock()

	path := strings.TrimPrefix(req.Path, Prefix)
	module, path, _ := strings.Cut(path, `.`)

	obj := r.m[module]
	if obj == nil {
		return nil, errors.New(`无此运行时配置。`)
	}

	u := config.NewUpdater(obj)
	u.MustApply(path, req.Yaml, func(path, value string) {})
	return &proto.SetConfigResponse{}, nil
}
