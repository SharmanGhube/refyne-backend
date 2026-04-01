package user

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

func SetupUserRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	userHandler := registry.User

	// Initialize rate limiter
	rateLimiter := middlewares.NewInMemoryRateLimiter(logging.GetComponentLogger("ratelimit"))

	// All user routes require authentication
	userGroup := router.Group("/user")
	userGroup.Use(middlewares.AuthMiddleware())
	userGroup.Use(rateLimiter.Middleware(middlewares.ProtectedEndpointLimit))
	{
		// Profile routes
		userGroup.GET("/profile", userHandler.GetProfile)
		userGroup.PUT("/profile", userHandler.UpdateProfile)

		// Settings routes
		userGroup.GET("/settings", userHandler.GetSettings)
		userGroup.PUT("/settings", userHandler.UpdateSettings)

		// Onboarding
		userGroup.POST("/onboarding", userHandler.CompleteOnboarding)

		// Account deletion
		userGroup.DELETE("/account", userHandler.DeleteAccount)
	}
}
