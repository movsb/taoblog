package protocols

import "github.com/gin-gonic/gin"

type AuthRequest struct {
	C *gin.Context
}

type AuthResponse struct {
	Success bool
}

type AuthLoginRequest struct {
	UserAgent string
	Username  string
	Password  string
}

type AuthLoginResponse struct {
	Success bool
	Cookie  string
}
