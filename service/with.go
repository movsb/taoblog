package service

import (
	"github.com/movsb/taoblog/modules/mailer"
	"github.com/movsb/taoblog/modules/notify"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
)

type With func(s *Service)

// 用于指定文章的附件存储。
func WithPostDataFileSystem(fsys theme_fs.FS) With {
	return func(s *Service) {
		s.postDataFS = fsys
	}
}

func WithNotifier(notifier notify.Notifier) With {
	return func(s *Service) {
		s.notifier = notifier
	}
}

// TODO 改成接口，像 Notify 一样。
func WithMailer(mailer *mailer.MailerLogger) With {
	return func(s *Service) {
		s.mailer = mailer
	}
}
