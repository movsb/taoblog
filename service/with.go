package service

import (
	"io/fs"
	"sync"

	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/cache"
	"github.com/movsb/taoblog/service/modules/storage"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
)

type With func(s *Service)

// 用于指定文章的附件存储。
func WithPostDataFileSystem(fsys *storage.SQLite) With {
	return func(s *Service) {
		s.mainStorage = fsys
		s.postDataFS = &WrappedPostFilesFileSystem{
			FS: fsys,
		}
	}
}

type WrappedPostFilesFileSystem struct {
	theme_fs.FS

	mounted sync.Map
}

func (fsys *WrappedPostFilesFileSystem) ForPost(id int) fs.FS {
	// 99.9% 的文章都是不走这条路的，会影响性能吗？
	if mfs, ok := fsys.mounted.Load(id); ok {
		return mfs.(fs.FS)
	}
	return fsys.FS.ForPost(id)
}

func (fsys *WrappedPostFilesFileSystem) Register(pid int, fileSystem fs.FS) {
	fsys.mounted.Store(pid, fileSystem)
}

func WithNotifier(notifier proto.NotifyServer) With {
	return func(s *Service) {
		s.notifier = notifier
	}
}

func WithCancel(cancel func()) With {
	return func(s *Service) {
		s.cancel = cancel
	}
}

func WithFileCache(cache *cache.FileCache) With {
	return func(s *Service) {
		s.fileCache = cache
	}
}
