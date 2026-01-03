package server

import (
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"google.golang.org/grpc"
)

type With func(s *Server)

// 请求节流器。
func WithRequestThrottler(throttler grpc.UnaryServerInterceptor) With {
	return func(s *Server) {
		s.throttler = throttler
		s.throttlerEnabled.Store(true)
	}
}

// 是否自动创建第一篇（自动生成的）文章。
func WithCreateFirstPost() With {
	return func(s *Server) {
		s.createFirstPost = true
	}
}

func WithTimezone(loc *time.Location) With {
	return func(s *Server) {
		s.initialTimezone = loc
	}
}

// 是否创建自动化 Git 仓库同步提交任务。
//
// 此任务自动把文章变更提交到 Git 仓库。
func WithGitSyncTask(b bool) With {
	return func(s *Server) {
		s.initGitSyncTask = b
	}
}

func WithBackupTasks(b bool) With {
	return func(s *Server) {
		s.initBackupTasks = b
	}
}

func WithRSS(b bool) With {
	return func(s *Server) {
		s.initRssTasks = b
	}
}

func WithMonitorDomain(b bool, initialDelay bool) With {
	return func(s *Server) {
		s.initMonitorDomain = b
		s.initMonitorDomainDelay = initialDelay
	}
}

func WithMonitorCerts(b bool) With {
	return func(s *Server) {
		s.initMonitorCerts = b
	}
}

// WithConfigOverride 设置一个函数来覆盖配置加载后的配置。
func WithConfigOverride(fn func(*config.Config)) With {
	return func(s *Server) {
		s.configOverride = fn
	}
}

func WithYearProgress() With {
	return func(s *Server) {
		s.initYearProgress = true
	}
}

func WithLiveCheck() With {
	return func(s *Server) {
		s.initLiveCheck = true
	}
}

func WithReviewerTask() With {
	return func(s *Server) {
		s.initReviewerTask = true
	}
}
