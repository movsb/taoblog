package service

import (
	"github.com/gin-gonic/gin"
	"github.com/movsb/taoblog/protocols"
)

type IAuth interface {
	AuthLogin(username string, password string) bool
	AuthCookie(c *gin.Context) bool
	AuthHeader(c *gin.Context) bool
	MakeCookie(userAgent string) string
}

func (s *ImplServer) Auth(in *protocols.AuthRequest) *protocols.AuthResponse {
	return &protocols.AuthResponse{
		Success: s.auth.AuthCookie(in.C) || s.auth.AuthHeader(in.C),
	}
}

func (s *ImplServer) AuthLogin(in *protocols.AuthLoginRequest) *protocols.AuthLoginResponse {
	ok := s.auth.AuthLogin(in.Username, in.Password)
	cookie := ""
	if ok {
		cookie = s.auth.MakeCookie(in.UserAgent)
	}
	return &protocols.AuthLoginResponse{
		Success: ok,
		Cookie:  cookie,
	}
}
