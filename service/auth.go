package service

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

type IAuth interface {
	AuthCookie(c *gin.Context) bool
	AuthHeader(c *gin.Context) bool
}

func (s *ImplServer) Auth(in *protocols.AuthRequest) *protocols.AuthResponse {
	return &protocols.AuthResponse{
		Success: s.auth.AuthCookie(in.C) || s.auth.AuthHeader(in.C),
	}
}
