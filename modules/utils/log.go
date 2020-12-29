package utils

import "go.uber.org/zap"

// InitTestLogger ...
func InitTestLogger() {
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}
