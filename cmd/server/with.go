package server

import (
	"time"

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
