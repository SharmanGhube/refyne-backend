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
		// ========== REGISTRATION ==========
		authGroup.POST("/register",
			rateLimiter.Middleware(middlewares.RegisterLimit),
			AuthHandler.Register)

		// ========== OTP LOGIN FLOW (Frontend-Expected Names) ==========
		authGroup.POST("/otp/send",
			rateLimiter.Middleware(middlewares.OTPRequestLimit),
			AuthHandler.RequestOTP)

		authGroup.POST("/otp/verify",
			rateLimiter.Middleware(middlewares.LoginLimit),
			AuthHandler.VerifyOTP)

		// ========== LEGACY OTP ENDPOINTS (Backward Compatibility) ==========
		authGroup.POST("/request-otp",
			rateLimiter.Middleware(middlewares.OTPRequestLimit),
			AuthHandler.RequestOTP)

		// ========== PASSWORD LOGIN ==========
		authGroup.POST("/login",
			rateLimiter.Middleware(middlewares.LoginLimit),
			AuthHandler.LoginWithPassword)

		// ========== TOKEN REFRESH ==========
		authGroup.POST("/refresh",
			rateLimiter.Middleware(middlewares.RefreshLimit),
			AuthHandler.RefreshToken)

		// ========== EMAIL VERIFICATION (Frontend-Expected Paths) ==========
		authGroup.POST("/verify/email",
			AuthHandler.VerifyAccount)

		authGroup.POST("/verify/email/resend",
			rateLimiter.Middleware(middlewares.VerificationResendLimit),
			AuthHandler.ResendVerification)

		// ========== LEGACY VERIFICATION ENDPOINTS (Backward Compatibility) ==========
		authGroup.POST("/verify",
			AuthHandler.VerifyAccount)

		authGroup.POST("/resend-verification",
			rateLimiter.Middleware(middlewares.VerificationResendLimit),
			AuthHandler.ResendVerification)

		// ========== PASSWORD RESET (Frontend-Expected Paths) ==========
		authGroup.POST("/password/reset/request",
			rateLimiter.Middleware(middlewares.PasswordResetLimit),
			AuthHandler.ForgotPassword)

		authGroup.POST("/password/reset/confirm",
			AuthHandler.ResetPassword)

		authGroup.POST("/password/reset/validate-token",
			AuthHandler.ValidateResetToken)

		// ========== LEGACY PASSWORD RESET ENDPOINTS (Backward Compatibility) ==========
		authGroup.POST("/forgot-password",
			rateLimiter.Middleware(middlewares.PasswordResetLimit),
			AuthHandler.ForgotPassword)

		authGroup.POST("/reset-password",
			AuthHandler.ResetPassword)

		authGroup.POST("/validate-reset-token",
			AuthHandler.ValidateResetToken)

		// ========== PROTECTED ROUTES (Authentication Required + Rate Limiting) ==========
		protected := authGroup.Group("")
		protected.Use(middlewares.AuthMiddleware())
		protected.Use(rateLimiter.Middleware(middlewares.ProtectedEndpointLimit))
		{
			protected.POST("/logout", AuthHandler.Logout)
			protected.POST("/logout-all", AuthHandler.LogoutAllDevices)
		}
	}
}
