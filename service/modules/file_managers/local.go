package file_managers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalFileManager implements FileManager
// on local file system.
type LocalFileManager struct {
	root string
}

// NewLocalFileManager creates a new LocalFileManager.
func NewLocalFileManager(root string) *LocalFileManager {
	wd, _ := os.Getwd()
	root = filepath.Join(wd, root)
	return &LocalFileManager{
		root: root,
	}
}

// Put implements IFileManager.Put
func (z *LocalFileManager) Put(pid int64, name string, r io.Reader) error {
	var err error
	root := filepath.Join(z.root, fmt.Sprint(pid))
	if err = os.MkdirAll(root, 0755); err != nil {
		return err
	}
	path := filepath.Join(root, name)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	if err != nil {
		return err
	}

	return nil
}

// Delete implements IFileManager.Delete
func (z *LocalFileManager) Delete(pid int64, name string) error {
	root := filepath.Join(z.root, fmt.Sprint(pid), name)
	return os.Remove(root)
}

// List implements IFileManager.List
func (z *LocalFileManager) List(pid int64) ([]string, error) {
	files := make([]string, 0)
	root := filepath.Join(z.root, fmt.Sprint(pid))

	err := filepath.Walk(root, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return filepath.SkipDir
		}

		if !info.IsDir() {
			files = append(files, info.Name())
		}

		return nil
	})

	return files, err
}
