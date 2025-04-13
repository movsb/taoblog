package service

import (
	"io/fs"

	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/cache"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
)

type With func(s *Service)

func WithThemeRootFileSystem(fsys fs.FS) With {
	return func(s *Service) {
		s.themeRootFS = fsys
	}
}

// 用于指定文章的附件存储。
func WithPostDataFileSystem(fsys theme_fs.FS) With {
	return func(s *Service) {
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

func WithFileCache(cache cache.FileCache) With {
	return func(s *Service) {
		s.fileCache = cache
	}
}
