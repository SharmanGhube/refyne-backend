package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

// SetupInstagramRoutes registers all Instagram routes
func SetupInstagramRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	handler := registry.Instagram
	if handler == nil {
		return
	}

	// Initialize rate limiter
	rateLimiter := middlewares.NewInMemoryRateLimiter(logging.GetComponentLogger("ratelimit"))

	// Public routes (no auth required)
	public := router.Group("")
	{
		// OAuth callback (redirect from Instagram)
		public.GET("/instagram/auth/callback", handler.OAuthCallback)

		// Webhook receiver (no auth required - signature verification handled in handler)
		public.GET("/instagram/webhooks", handler.HandleWebhook)
		public.POST("/instagram/webhooks", handler.HandleWebhook)
	}

	// Protected routes (auth required)
	protected := router.Group("")
	protected.Use(middlewares.AuthMiddleware())
	protected.Use(rateLimiter.Middleware(middlewares.ProtectedEndpointLimit))
	{
		// OAuth connection
		protected.POST("/instagram/auth/connect", handler.ConnectAccount)
		protected.POST("/instagram/auth/disconnect", handler.DisconnectAccount)

		// Account management
		protected.GET("/instagram/accounts", handler.ListAccounts)
		protected.GET("/instagram/accounts/:id", handler.GetAccount)

		// Media management
		protected.GET("/instagram/media", handler.GetMedia)
		protected.GET("/instagram/media/:id", handler.GetMediaByID)
		protected.GET("/instagram/media/:id/ai", handler.GetMediaRecommendations)

		// Analytics
		protected.GET("/instagram/analytics", handler.GetAccountAnalytics)
		protected.GET("/instagram/analytics/media", handler.GetMediaAnalytics)
		protected.GET("/instagram/analytics/trends", handler.GetAnalyticsTrends)

		// AI features
		protected.POST("/instagram/ai/caption-suggest", handler.GenerateCaptions)
		protected.POST("/instagram/ai/hashtag-suggest", handler.GenerateHashtags)
		protected.GET("/instagram/ai/posting-time", handler.GetPostingStrategy)

		// Manual operations
		protected.POST("/instagram/media/sync", handler.ManualSync)
		protected.POST("/instagram/media/analyze", handler.ManualAnalyze)
	}
}
