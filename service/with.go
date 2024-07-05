package service

import (
	"github.com/movsb/taoblog/modules/notify"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"google.golang.org/grpc"
)

type With func(s *Service)

// 用于指定文章的附件存储。
func WithPostDataFileSystem(fsys theme_fs.FS) With {
	return func(s *Service) {
		s.postDataFS = fsys
	}
}

func WithInstantNotifier(instantNotifier notify.InstantNotifier) With {
	return func(s *Service) {
		s.instantNotifier = instantNotifier
	}
}

// 请求节流器。
func WithRequestThrottler(throttler grpc.UnaryServerInterceptor) With {
	return func(s *Service) {
		s.throttler = throttler
		s.throttlerEnabled.Store(true)
	}
}
