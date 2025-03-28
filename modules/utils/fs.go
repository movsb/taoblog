package utils

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
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
				log.Println(err)
				time.Sleep(time.Second)
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
		if fp, err := layer.Open(name); err == nil {
			return fp, nil
		}
	}
	return nil, &fs.PathError{Op: `open`, Path: name, Err: fs.ErrNotExist}
}

// 前面参数的层先被打开（也即：前面的是上层）。
func NewOverlayFS(layers ...fs.FS) fs.FS {
	return &_OverlayFS{layers: layers}
}
