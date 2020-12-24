package local

import (
	"os"
	"path/filepath"
)

// walk returns the file list for a directory.
// Directories are omitted.
// Returned paths are related to dir.
func walk(dir string) (files []string, err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			files = append(files, rel)

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return files, nil
}
