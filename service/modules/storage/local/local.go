package local

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/movsb/taoblog/service/modules/storage"
)

// Local stores files on local file system.
type Local struct {
	root string
}

// NewLocal ...
func NewLocal(root string) (*Local, error) {
	if err := os.MkdirAll(root, 0755); err != nil {
		return nil, err
	}
	return &Local{root: root}, nil
}

var _ storage.Store = &Local{}

func (l *Local) path(id int64, path string, createDir bool) (string, error) {
	dir := filepath.Join(l.root, fmt.Sprint(id))
	if createDir {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", err
		}
	}
	path = filepath.Join(dir, path)
	return path, nil
}

// List ...
func (l *Local) List(id int64) ([]string, error) {
	path, err := l.path(id, `.`, false)
	if err != nil {
		return nil, err
	}
	return walk(path)
}

// Open ...
func (l *Local) Open(id int64, path string) (storage.File, error) {
	path, err := l.path(id, path, false)
	if err != nil {
		return nil, err
	}
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

// Create ...
func (l *Local) Create(id int64, path string) (storage.File, error) {
	path, err := l.path(id, path, true)
	if err != nil {
		return nil, err
	}
	fp, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return fp, nil
}

// Remove ...
func (l *Local) Remove(id int64, path string) error {
	path, err := l.path(id, path, false)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
