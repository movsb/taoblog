package service

import (
	"time"
)

// 控制 RSS 等的输出稳定。
func (s *Service) TestingSetLastPostedAt(t time.Time) {
	s.updateLastPostTime(t)
}

func (s *Service) TestingSetTimezone(t *time.Location) {
	s.timeLocation = t
}

func (s *Service) TestingSetHTTPAddr(u string) {
	s.getHome = func() string { return u }
}
