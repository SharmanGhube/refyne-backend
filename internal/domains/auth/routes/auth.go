package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

func SetupAuthRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	AuthHandler := registry.Auth

	// Initialize rate limiter
	rateLimiter := middlewares.NewInMemoryRateLimiter(logging.GetComponentLogger("ratelimit"))

	authGroup := router.Group("/auth")
	{
		// Public routes with rate limiting
		authGroup.POST("/register",
			rateLimiter.Middleware(middlewares.RegisterLimit),
			AuthHandler.Register)

		authGroup.POST("/request-otp",
			rateLimiter.Middleware(middlewares.OTPRequestLimit),
			AuthHandler.RequestOTP)

		authGroup.POST("/login",
			rateLimiter.Middleware(middlewares.LoginLimit),
			AuthHandler.VerifyOTP)

		authGroup.POST("/refresh",
			rateLimiter.Middleware(middlewares.RefreshLimit),
			AuthHandler.RefreshToken)

		// Email verification routes (with rate limiting)
		authGroup.POST("/verify", AuthHandler.VerifyAccount)

		authGroup.POST("/resend-verification",
			rateLimiter.Middleware(middlewares.VerificationResendLimit),
			AuthHandler.ResendVerification)

		// Password reset routes (with rate limiting)
		authGroup.POST("/forgot-password",
			rateLimiter.Middleware(middlewares.PasswordResetLimit),
			AuthHandler.ForgotPassword)

		authGroup.POST("/reset-password", AuthHandler.ResetPassword)
		authGroup.POST("/validate-reset-token", AuthHandler.ValidateResetToken)

		// Protected routes (authentication required + rate limiting)
		protected := authGroup.Group("")
		protected.Use(middlewares.AuthMiddleware())
		protected.Use(rateLimiter.Middleware(middlewares.ProtectedEndpointLimit))
		{
			protected.POST("/logout", AuthHandler.Logout)
			protected.POST("/logout-all", AuthHandler.LogoutAllDevices)
		}
	}
}
