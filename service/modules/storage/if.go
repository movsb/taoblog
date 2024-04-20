package storage

import (
	"fmt"
	"io"
	fspkg "io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
)

// File ...
type File interface {
	io.Reader
	io.Writer
	io.Seeker
	io.Closer
	Stat() (os.FileInfo, error)
}

// 文件子系统接口。
// 所谓“子”，就是针对某篇文章。
// TODO 应该用标准的接口。
type FileSystem interface {
	ListFiles() ([]*protocols.FileSpec, error)
	DeleteFile(path string) error
	OpenFile(path string) (File, error)
	WriteFile(spec *protocols.FileSpec, r io.Reader) error
	Resolve(path string) string
}

// 针对某篇文章的文件系统实现类。
// 目录结构：配置的文章附件根目录/文章编号/附件路径。
// TODO 改成全局一个实例统一管理所有文章的文件。
type Local struct {
	root string
	id   int64
	dir  string
}

var _ FileSystem = (*Local)(nil)

func NewLocal(root string, id int64) *Local {
	return &Local{
		root: root,
		id:   id,
		dir:  filepath.Join(root, fmt.Sprint(id)),
	}
}

func (fs *Local) pathOf(path string) string {
	if path == "" {
		panic("path cannot be empty")
	}
	return filepath.Join(fs.dir, filepath.Clean(path))
}

func (fs *Local) ListFiles() ([]*protocols.FileSpec, error) {
	files, err := utils.ListFiles(fs.dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return nil, err
	}
	return files, nil
}

func (fs *Local) DeleteFile(path string) error {
	path = filepath.Clean(path)
	path = fs.pathOf(path)
	return os.Remove(path)
}

func (fs *Local) OpenFile(path string) (File, error) {
	path = fs.pathOf(path)
	return os.Open(path)
}

func (fs *Local) WriteFile(spec *protocols.FileSpec, r io.Reader) error {
	path := fs.pathOf(spec.Path)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return utils.WriteFile(path, fspkg.FileMode(spec.Mode), time.Unix(int64(spec.Time), 0), int64(spec.Size), r)
}

func (fs *Local) Resolve(path string) string {
	return fs.pathOf(path)
}
