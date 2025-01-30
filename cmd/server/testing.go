package server

// 测试的时候需要禁用限流器。
func (s *Server) TestEnableRequestThrottler(enable bool) {
	s.throttlerEnabled.Store(enable)
}
