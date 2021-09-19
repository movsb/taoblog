package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/movsb/taoblog/protocols"
)

func ListBackupFiles(dir string) ([]*protocols.BackupFileSpec, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if len(dir) == 1 {
		return nil, fmt.Errorf(`dir cannot be root`)
	}

	files := make([]*protocols.BackupFileSpec, 0, 1024)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() && !info.Mode().IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		file := &protocols.BackupFileSpec{
			Path: rel,
			Mode: uint32(info.Mode()),
			Size: uint32(info.Size()),
			Time: uint32(info.ModTime().Unix()),
		}

		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}
