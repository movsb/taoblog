package roots

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type File struct {
	Path string `json:"path" yaml:"path"`
	Time int64  `json:"time" yaml:"time"`
	Type string `json:"type" yaml:"type"`
	Data []byte `json:"data" yaml:"data"`

	io.ReadSeeker `json:"-" yaml:"-"`
}

var _ interface {
	fs.File
} = (*File)(nil)

func (f *File) Stat() (fs.FileInfo, error) {
	return &Info{f: f}, nil
}
func (f *File) Close() error { return nil }

type Info struct {
	f *File
}

var _ interface {
	fs.FileInfo
} = (*Info)(nil)

func (f *Info) Name() string {
	return filepath.Base(f.f.Path)
}
func (f *Info) Size() int64 {
	return int64(len(f.f.Data))
}
func (f *Info) Mode() fs.FileMode {
	return 0 | 0600
}
func (f *Info) ModTime() time.Time {
	return time.Unix(f.f.Time, 0)
}
func (f *Info) IsDir() bool {
	return f.Mode().IsDir()
}
func (f *Info) Sys() any {
	return nil
}

type Root struct {
	store          utils.PluginStorage
	mux            utils.HTTPMux
	registeredDirs map[string]struct{}
	lock           sync.Mutex
}

func New(p utils.PluginStorage, mux utils.HTTPMux) *Root {
	r := &Root{
		store:          p,
		mux:            mux,
		registeredDirs: map[string]struct{}{},
	}
	go r.load()
	return r
}

func (r *Root) load() {
	r.store.Range(func(key string) {
		if key[0] == '/' {
			r.registerDir(key)
		}
	})
}

func (r *Root) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	http.ServeFileFS(w, req, r, req.URL.Path)
}

func (r *Root) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: `open`, Path: name, Err: errors.New(`invalid path`)}
	}
	value, err := r.store.GetString(`/` + name)
	if err != nil {
		if taorm.IsNotFoundError(err) {
			return nil, os.ErrNotExist
		}
		return nil, err
	}
	var file File
	if err := json.Unmarshal([]byte(value), &file); err != nil {
		return nil, err
	}
	file.ReadSeeker = bytes.NewReader(file.Data)
	return &file, nil
}

func (r *Root) GetConfig(ctx context.Context, req *proto.GetConfigRequest) (*proto.GetConfigResponse, error) {
	f, err := r.Open(req.Path[1:])
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return &proto.GetConfigResponse{
		Yaml: string(utils.Must1(io.ReadAll(f))),
	}, nil
}

func (r *Root) SetConfig(ctx context.Context, req *proto.SetConfigRequest) (*proto.SetConfigResponse, error) {
	if len(req.Yaml) > 1<<20 {
		return nil, status.Error(codes.InvalidArgument, `文件太大。`)
	}
	if strings.HasSuffix(req.Path, `/`) {
		return nil, status.Errorf(codes.InvalidArgument, `文件名错误：%s`, req.Path)
	}

	file := File{
		Path: req.Path,
		Time: time.Now().Unix(),
		Type: ``,
		Data: []byte(req.Yaml),
	}
	b, err := json.Marshal(file)
	if err != nil {
		return nil, err
	}

	if err := r.store.SetString(file.Path, string(b)); err != nil {
		return nil, err
	}

	r.registerDir(file.Path)

	return &proto.SetConfigResponse{}, nil
}

// NOTE: 只注册了第一层目录，可能不精确。
func (r *Root) registerDir(path string) {
	// /.well-known/security.txt → /.well-known/
	comps := strings.Split(path, `/`)
	var group string
	if len(comps) <= 2 {
		group = path
	} else {
		group = fmt.Sprintf(`/%s/`, comps[1])
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	if _, ok := r.registeredDirs[group]; !ok {
		escaped := (&url.URL{Path: group}).EscapedPath()
		// TODO: 可能 panic。
		r.mux.Handle(escaped, r)
		r.registeredDirs[group] = struct{}{}
		log.Println(`注册动态文件：`, group)
	}
}
