package service

import (
	"github.com/movsb/taoblog/service/modules/pingback/xmlrpc"
	"go.uber.org/zap"
)

// Pingback is the pingback service handler.
func (s *Service) Pingback(w xmlrpc.ResponseWriter, source string, target string) {
	zap.L().Info(`pingback`, zap.String("source", source), zap.String("target", target))
}
