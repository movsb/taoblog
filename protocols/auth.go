package protocols

import "github.com/gin-gonic/gin"

type AuthRequest struct {
	C *gin.Context
}

type AuthResponse struct {
	Success bool
}
