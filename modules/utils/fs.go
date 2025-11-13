package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/movsb/taoblog/protocols/go/proto"
)

type WatchFS interface {
	Watch() (<-chan fsnotify.Event, func(), error)
}

// 扩展了 os.DirFS，支持 SubFS 和 WatchFS。
type OSDirFS struct {
	root string
	fs.FS
}

var _ interface {
	fs.FS
	WatchFS
	fs.SubFS
} = (*OSDirFS)(nil)

// 扩展了 os.DirFS，支持 SubFS 和 WatchFS。
func NewOSDirFS(root string) fs.FS {
	return &OSDirFS{
		root: root,
		FS:   os.DirFS(root),
	}
}

func (fsys *OSDirFS) Root() string {
	return fsys.root
}

func (fsys *OSDirFS) Sub(dir string) (fs.FS, error) {
	if !fs.ValidPath(dir) {
		return nil, &fs.PathError{
			Op:   `sub`,
			Path: dir,
			Err:  errors.New(`invalid path`),
		}
	}
	if dir == `.` {
		return fsys, nil
	}
	return NewOSDirFS(path.Join(fsys.root, dir)), nil
}

// TODO 改成通过 context 退出，而不是返回 close 方法。
func (fsys *OSDirFS) Watch() (<-chan fsnotify.Event, func(), error) {
	if _, err := fsys.Open("."); err != nil {
		panic(fmt.Sprintf(`err: %v, cwd: %v, root: %v`, err, Must1(os.Getwd()), fsys.root))
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	ch := make(chan fsnotify.Event)

	go func() {
		for {
			select {
			case err := <-watcher.Errors:
				log.Println("Watch Error:", err, fsys.root)
				return
			case event := <-watcher.Events:
				// log.Println(event)
				ch <- event
			}
		}
	}()

	if err := watcher.Add(fsys.root); err != nil {
		return nil, nil, err
	}

	return ch, func() { watcher.Close() }, nil
}

// 作为对 fs.FS 的补充。
// 官方标准化了我就删。
type DeleteFS interface {
	fs.FS
	Delete(name string) error
}

func Delete(fsys fs.FS, name string) error {
	if dfs, ok := fsys.(DeleteFS); ok {
		return dfs.Delete(name)
	}
	return errors.New(`fs.Delete: unimplemented`)
}

type WriteFS interface {
	Write(spec *proto.FileSpec, r io.Reader) error
}

func Write(fsys fs.FS, spec *proto.FileSpec, r io.Reader) error {
	if wfs, ok := fsys.(WriteFS); ok {
		return wfs.Write(spec, r)
	}
	return errors.New(`fs.Write: unimplemented`)
}

type ListFilesFS interface {
	ListFiles() ([]*proto.FileSpec, error)
}

func ListFiles(fsys fs.FS) ([]*proto.FileSpec, error) {
	if lfs, ok := fsys.(ListFilesFS); ok {
		return lfs.ListFiles()
	}
	return nil, errors.New(`fs.ListFiles: unimplemented`)
}

////////////////////////////////////////////////////////////////////////////////

type _OverlayFS struct {
	layers []fs.FS
}

func (fsys *_OverlayFS) Open(name string) (fs.File, error) {
	for _, layer := range fsys.layers {
		if layer == nil {
			continue
		}
		fp, err := layer.Open(name)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			return nil, err
		}
		return fp, nil
	}
	return nil, &fs.PathError{Op: `open`, Path: name, Err: fs.ErrNotExist}
}

// 前面参数的层先被打开（也即：前面的是上层）。
// nil 会被忽略。
func NewOverlayFS(layers ...fs.FS) fs.FS {
	return &_OverlayFS{layers: layers}
}

type StringFile struct {
	Name string `json:"name" yaml:"name"`
	Time int64  `json:"time" yaml:"time"`
	Data []byte `json:"data" yaml:"data"`

	io.ReadSeeker `json:"-" yaml:"-"`
}

func NewStringFile(name string, mod time.Time, data []byte) *StringFile {
	return &StringFile{
		Name:       name,
		Time:       mod.Unix(),
		Data:       data,
		ReadSeeker: bytes.NewReader(data),
	}
}

var _ interface {
	fs.File
} = (*StringFile)(nil)

func (f *StringFile) Stat() (fs.FileInfo, error) {
	return &_FileInfo{f: f}, nil
}
func (f *StringFile) Close() error { return nil }

type _FileInfo struct {
	f *StringFile
}

var _ interface {
	fs.FileInfo
} = (*_FileInfo)(nil)

func (f *_FileInfo) Name() string {
	return filepath.Base(f.f.Name)
}
func (f *_FileInfo) Size() int64 {
	return int64(len(f.f.Data))
}
func (f *_FileInfo) Mode() fs.FileMode {
	return 0 | 0600
}
func (f *_FileInfo) ModTime() time.Time {
	return time.Unix(f.f.Time, 0)
}
func (f *_FileInfo) IsDir() bool {
	return f.Mode().IsDir()
}
func (f *_FileInfo) Sys() any {
	return nil
}
