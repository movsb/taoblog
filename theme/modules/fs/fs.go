package theme_fs

import (
	"embed"
	"io/fs"
)

// 用于对外隐藏存储结构。
// TODO：也应该用于本地备份时的文件系统，方便恢复。
type FS interface {
	// 用于整个备份。
	Root() (fs.FS, error)
	// 针对单篇文章/评论的文件系统。
	ForPost(id int) (fs.FS, error)
}

type Empty struct{}

var empty embed.FS

func (Empty) Root() (fs.FS, error) {
	return empty, nil
}

func (Empty) ForPost(id int) (fs.FS, error) {
	return empty, nil
}
