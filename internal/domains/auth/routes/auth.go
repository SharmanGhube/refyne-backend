package auth

import (
	"github.com/gin-gonic/gin"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
)

func SetupAuthRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	AuthHandler := registry.Auth
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", AuthHandler.Register)
		authGroup.POST("/request-otp", AuthHandler.RequestOTP)
		authGroup.POST("/login", AuthHandler.VerifyOTP) // Now this handles OTP verification
		authGroup.POST("/refresh", AuthHandler.RefreshToken)
		authGroup.POST("/verify", AuthHandler.VerifyAccount)
	}
}
