package service

import (
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/cache"
	"github.com/movsb/taoblog/service/modules/storage"
)

type With func(s *Service)

// 用于指定文章的附件存储。
func WithPostDataFileSystem(fsys *storage.SQLite) With {
	return func(s *Service) {
		s.mainStorage = fsys
		s.postDataFS = fsys
	}
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
