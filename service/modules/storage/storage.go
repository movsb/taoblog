package storage

import (
	"fmt"
	"io"
	"io/fs"
	fspkg "io/fs"
	"path/filepath"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/spf13/afero"
)

type Storage struct {
	root afero.Fs
}

var _ interface {
	fspkg.FS
	fspkg.StatFS
	fspkg.ReadDirFS
	fspkg.SubFS
	utils.DeleteFS
	utils.WriteFS

	theme_fs.FS
} = (*Storage)(nil)

func NewLocal(root string) *Storage {
	l := &Storage{
		root: afero.NewBasePathFs(afero.NewOsFs(), root),
	}
	return l
}

func NewMemory() *Storage {
	return &Storage{
		root: afero.NewMemMapFs(),
	}
}

type _FsOpen afero.Afero

func (fs *_FsOpen) Open(name string) (fs.File, error) {
	return fs.Fs.Open(name)
}

func (fs *Storage) Root() (fspkg.FS, error) {
	return &_FsOpen{Fs: fs.root}, nil
}

func (fs *Storage) ForPost(id int) (fspkg.FS, error) {
	return fs.Sub(fmt.Sprint(id))
}

func (fs *Storage) pathOf(path string) (string, error) {
	if !fspkg.ValidPath(path) {
		return "", fmt.Errorf(`invalid fs path: %q` + path)
	}
	// ValidPath 会检查是可能 .. 到上级。
	return path, nil
}

func (fs *Storage) Open(name string) (fspkg.File, error) {
	path, err := fs.pathOf(name)
	if err != nil {
		return nil, err
	}
	return fs.root.Open(path)
}

func (fs *Storage) Delete(name string) error {
	path, err := fs.pathOf(name)
	if err != nil {
		return err
	}
	return fs.root.Remove(path)
}

func (fs *Storage) Sub(dir string) (fspkg.FS, error) {
	path, err := fs.pathOf(dir)
	if err != nil {
		return nil, err
	}
	// TODO: afero 有个 bug：https://github.com/spf13/afero/issues/428
	return &Storage{root: afero.NewBasePathFs(fs.root, "/"+path)}, nil
}

// 会自动创建不存在的目录。
func (fs *Storage) Write(spec *proto.FileSpec, r io.Reader) error {
	// NOTE：实际上并没有什么用/并不关心。只是想看看有没有恶意上传😏。
	// mode := fspkg.FileMode(spec.Mode)
	// if strings.Contains(mode.Perm().String(), `x`) {
	// 	return fmt.Errorf(`不允许上传带可执行权限位的文件。`)
	// }

	path, err := fs.pathOf(spec.Path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := fs.root.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return utils.WriteFile(fs.root, path, fspkg.FileMode(spec.Mode), time.Unix(int64(spec.Time), 0), int64(spec.Size), r)
}

func (fs *Storage) Stat(name string) (fspkg.FileInfo, error) {
	path, err := fs.pathOf(name)
	if err != nil {
		return nil, err
	}
	return fs.root.Stat(path)
}

func (fs *Storage) ReadDir(name string) ([]fspkg.DirEntry, error) {
	path, err := fs.pathOf(name)
	if err != nil {
		return nil, err
	}
	fi, err := afero.ReadDir(fs.root, path)
	if err != nil {
		return nil, err
	}

	return utils.Map(fi, func(i fspkg.FileInfo) fspkg.DirEntry {
		return fspkg.FileInfoToDirEntry(i)
	}), nil
}
