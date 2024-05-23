package storage

import (
	"fmt"
	"io"
	"io/fs"
	fspkg "io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	proto "github.com/movsb/taoblog/protocols"
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
	ListFiles() ([]*proto.FileSpec, error)
	DeleteFile(path string) error
	OpenFile(path string) (File, error)
	WriteFile(spec *proto.FileSpec, r io.Reader) error
	Resolve(path string) string
}

// 针对某篇文章的文件系统实现类。
// 目录结构：配置的文章附件根目录/文章编号/附件路径。
// TODO 改成全局一个实例统一管理所有文章的文件。
type Local struct {
	root string
	dir  string

	maxFileSize int32
}

var _ interface {
	FileSystem
	fs.FS
} = (*Local)(nil)

type Option func(*Local)

func WithMaxFileSize(size int32) Option {
	return func(l *Local) {
		l.maxFileSize = size
	}
}

func NewLocal(root string, sub string, options ...Option) *Local {
	l := &Local{
		root: root,
		dir:  filepath.Join(root, sub),
	}
	for _, opt := range options {
		opt(l)
	}
	return l
}

func (fs *Local) pathOf(path string) string {
	if path == "" {
		panic("path cannot be empty")
	}
	return filepath.Join(fs.dir, filepath.Clean(path))
}

func (fs *Local) ListFiles() ([]*proto.FileSpec, error) {
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

func (fs *Local) Open(name string) (fspkg.File, error) {
	return fs.OpenFile(name)

}

func (fs *Local) WriteFile(spec *proto.FileSpec, r io.Reader) error {
	if fs.maxFileSize > 0 && spec.Size > uint32(fs.maxFileSize) {
		return fmt.Errorf(`文件太大（允许大小：%v 字节）。`, fs.maxFileSize)
	}
	// NOTE：实际上并没有什么用/并不关心。只是想看看有没有恶意上传😏。
	mode := fspkg.FileMode(spec.Mode)
	if strings.Contains(mode.Perm().String(), `x`) {
		return fmt.Errorf(`不允许上传带可执行权限位的文件。`)
	}

	path := fs.pathOf(spec.Path)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return utils.WriteFile(path, mode, time.Unix(int64(spec.Time), 0), int64(spec.Size), r)
}

func (fs *Local) Resolve(path string) string {
	return fs.pathOf(path)
}
