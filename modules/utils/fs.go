package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/movsb/taoblog/protocols"
)

// walk returns the file list for a directory.
// Directories are omitted.
// Returned paths are related to dir.
// 返回的所有路径都是相对于 dir 而言的。
func ListFiles(dir string) ([]*protocols.FileSpec, error) {
	bfs, err := ListBackupFiles(dir)
	if err != nil {
		return nil, err
	}
	fs := make([]*protocols.FileSpec, 0, len(bfs))
	for _, f := range bfs {
		fs = append(fs, &protocols.FileSpec{
			Path: f.Path,
			Mode: f.Mode,
			Size: f.Size,
			Time: f.Time,
		})
	}
	return fs, nil
}

// Deprecated. 用 ListFiles。
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
		if !info.Mode().IsRegular() || info.Mode().IsDir() {
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

// 安全写文件。
// TODO 不应该引入专用的 FileSpec 定义。
// path 中包含的目录必须存在，否则失败。
// TODO 没移除失败的文件。
// NOTE：安全写：先写临时文件，再移动过去。临时文件写在目标目录，不存在跨设备移动文件的问题。
// NOTE：如果 r 超过 size，会报错。
func WriteFile(path string, mode fs.FileMode, modified time.Time, size int64, r io.Reader) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), `taoblog-*`)
	if err != nil {
		return err
	}

	if n, err := io.Copy(tmp, io.LimitReader(r, size+1)); err != nil || n != size {
		return fmt.Errorf(`write error: %d %v`, n, err)
	}

	if err := tmp.Chmod(mode); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Chtimes(tmp.Name(), modified, modified); err != nil {
		return err
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		return err
	}

	return nil
}
