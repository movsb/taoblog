package utils

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/movsb/taoblog/protocols/go/proto"
)

// walk returns the file list for a directory.
// Directories are omitted.
// Returned paths are related to dir.
// 返回的所有路径都是相对于 dir 而言的。
func ListFiles(fsys fs.FS, dir string) ([]*proto.FileSpec, error) {
	bfs, err := listBackupFiles(fsys, dir)
	if err != nil {
		return nil, err
	}
	fs := make([]*proto.FileSpec, 0, len(bfs))
	for _, f := range bfs {
		fs = append(fs, &proto.FileSpec{
			Path: f.Path,
			Mode: f.Mode,
			Size: f.Size,
			Time: f.Time,
			Type: mime.TypeByExtension(filepath.Ext(f.Path)),
		})
	}
	return fs, nil
}

// Deprecated. 用 ListFiles。
func listBackupFiles(fsys fs.FS, dir string) ([]*proto.FileSpec, error) {
	files := make([]*proto.FileSpec, 0, 32)

	err := fs.WalkDir(fsys, dir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		info2, err := info.Info()
		if err != nil {
			return err
		}

		file := &proto.FileSpec{
			Path: rel,
			Mode: uint32(info2.Mode().Perm()),
			Size: uint32(info2.Size()),
			Time: uint32(info2.ModTime().Unix()),
		}

		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

type FsWithChangeNotify interface {
	fs.FS
	Changed() <-chan fsnotify.Event
}

type DirFSWithNotify struct {
	root string
	fs.FS
	ch chan fsnotify.Event
}

var _ FsWithChangeNotify = (*DirFSWithNotify)(nil)

func NewDirFSWithNotify(root string) fs.FS {
	l := &DirFSWithNotify{
		root: root,
		FS:   os.DirFS(root),
	}
	l.ch = l.watch()
	return l
}

func (l *DirFSWithNotify) Changed() <-chan fsnotify.Event {
	return l.ch
}

func (l *DirFSWithNotify) watch() chan fsnotify.Event {
	if _, err := l.FS.Open("."); err != nil {
		panic(fmt.Sprintf(`err: %v, cwd: %v, root: %v`, err, Must1(os.Getwd()), l.root))
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return nil
	}
	// defer watcher.Close()

	ch := make(chan fsnotify.Event)

	go func() {
		for {
			select {
			case err := <-watcher.Errors:
				log.Println(err)
				return
			case event := <-watcher.Events:
				// log.Println(event)
				ch <- event
			}
		}
	}()

	if err := watcher.Add(l.root); err != nil {
		panic(err)
	} else {
		log.Println(`Started watching`, l.root)
	}

	return ch
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
