package dir

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var Root string

// 必须在项目根目录下启动程序，否则结果一定是错误的。
func init() {
	rootDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	Root = rootDir
}

type Dir string

func (d Dir) Join(components ...string) string {
	return filepath.Join(append([]string{string(d)}, components...)...)
}

func (d Dir) FS() fs.FS {
	return os.DirFS(string(d))
}

func SourceRelativeDir() Dir {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic(`无法获取路径。`)
	}
	if Root == "" {
		panic(`没有设置根目录。`)
	}
	dir := filepath.Dir(file)
	s := strings.TrimPrefix(dir, Root)
	s = strings.TrimPrefix(s, "/")
	if s == "" {
		s = "."
	}
	return Dir(s)
}
