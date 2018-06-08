package file_managers

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/movsb/alioss/alioss"
)

// AliossFileManager implements FileManager
// on local file system.
type AliossFileManager struct {
	root string
	oss  *alioss.Client
}

// NewAliossFileManager creates a new AliossFileManager.
func NewAliossFileManager(bucket, location, key, secret string) *AliossFileManager {
	return &AliossFileManager{
		root: "files",
		oss:  alioss.NewClient(bucket, location, key, secret),
	}
}

// Put implements IFileManager.Put
func (z *AliossFileManager) Put(pid int64, name string, r io.Reader) error {
	path := filepath.Join(z.root, fmt.Sprint(pid), name)
	return z.oss.PutFile(path, ioutil.NopCloser(r))
}

// Delete implements IFileManager.Delete
func (z *AliossFileManager) Delete(pid int64, name string) error {
	path := filepath.Join(z.root, fmt.Sprint(pid), name)
	return z.oss.DeleteObject(path)
}

// List implements IFileManager.List
func (z *AliossFileManager) List(pid int64) ([]string, error) {
	path := filepath.Join(z.root, fmt.Sprint(pid))
	files, _, err := z.oss.ListFolder(path, false)
	f := make([]string, 0, len(files))
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		f = append(f, file.Key)
	}
	return f, nil
}
