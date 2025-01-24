package server

import "google.golang.org/grpc"

type With func(s *Server)

// 请求节流器。
func WithRequestThrottler(throttler grpc.UnaryServerInterceptor) With {
	return func(s *Server) {
		s.throttler = throttler
		s.throttlerEnabled.Store(true)
	}
}
