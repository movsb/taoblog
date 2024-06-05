package storage

import (
	"fmt"
	"io"
	fspkg "io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
)

type Local struct {
	root string
}

var _ interface {
	fspkg.FS
	fspkg.StatFS
	fspkg.ReadDirFS
	fspkg.SubFS
	utils.DeleteFS
	utils.WriteFS

	theme_fs.FS
} = (*Local)(nil)

func NewLocal(root string) *Local {
	l := &Local{
		root: root,
	}
	return l
}

func (fs *Local) Root() (fspkg.FS, error) {
	return os.DirFS(fs.root), nil
}

func (fs *Local) ForPost(id int) (fspkg.FS, error) {
	return fs.Sub(fmt.Sprint(id))
}

func (fs *Local) pathOf(path string) (string, error) {
	if !fspkg.ValidPath(path) {
		return "", fmt.Errorf(`invalid fs path: %q` + path)
	}
	// ValidPath 会检查是可能 .. 到上级。
	return filepath.Join(fs.root, path), nil
}

func (fs *Local) Open(name string) (fspkg.File, error) {
	path, err := fs.pathOf(name)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (fs *Local) Delete(name string) error {
	path, err := fs.pathOf(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (fs *Local) Sub(dir string) (fspkg.FS, error) {
	path, err := fs.pathOf(dir)
	if err != nil {
		return nil, err
	}

	return NewLocal(path), nil
}

// 会自动创建不存在的目录。
func (fs *Local) Write(spec *proto.FileSpec, r io.Reader) error {
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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return utils.WriteFile(path, fspkg.FileMode(spec.Mode), time.Unix(int64(spec.Time), 0), int64(spec.Size), r)
}

func (fs *Local) Stat(name string) (fspkg.FileInfo, error) {
	path, err := fs.pathOf(name)
	if err != nil {
		return nil, err
	}
	return os.Stat(path)
}

func (fs *Local) ReadDir(name string) ([]fspkg.DirEntry, error) {
	path, err := fs.pathOf(name)
	if err != nil {
		return nil, err
	}
	return os.ReadDir(path)
}
