package theme_fs

import (
	"embed"
	"io/fs"
	"time"

	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
)

// 用于对外隐藏存储结构。
// TODO：也应该用于本地备份时的文件系统，方便恢复。
type FS interface {
	// 用于整个备份。
	AllFiles() (map[int][]*proto.FileSpec, error)
	// 针对单篇文章/评论的文件系统。
	ForPost(id int) fs.FS
}

type Empty struct{}

var empty embed.FS

func (Empty) AllFiles() (map[int][]*proto.FileSpec, error) {
	return nil, nil
}

func (Empty) ForPost(id int) fs.FS {
	return empty
}

type FileURLGetter interface {
	GetFileURL(post *proto.Post, file *models.File, ttl time.Duration) (string, string, bool, error)
}
