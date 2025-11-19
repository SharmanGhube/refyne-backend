package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
)

func SetupAuthRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	AuthHandler := registry.Auth
	authGroup := router.Group("/auth")
	{
		// Public routes (no authentication required)
		authGroup.POST("/register", AuthHandler.Register)
		authGroup.POST("/request-otp", AuthHandler.RequestOTP)
		authGroup.POST("/login", AuthHandler.VerifyOTP) // OTP verification and login
		authGroup.POST("/refresh", AuthHandler.RefreshToken)
		authGroup.POST("/verify", AuthHandler.VerifyAccount)

		// Protected routes (authentication required)
		protected := authGroup.Group("")
		protected.Use(middlewares.AuthMiddleware())
		{
			protected.POST("/logout", AuthHandler.Logout)
			protected.POST("/logout-all", AuthHandler.LogoutAllDevices)
		}
	}
}
