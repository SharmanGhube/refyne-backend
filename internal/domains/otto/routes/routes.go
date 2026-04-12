package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

// SetupOttoRoutes registers all Otto AI assistant routes
func SetupOttoRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	handler := registry.Otto
	if handler == nil {
		return
	}

	logger := logging.GetComponentLogger("otto-routes")
	logger.Info("Setting up Otto AI assistant routes")

	// Initialize rate limiter
	rateLimiter := middlewares.NewInMemoryRateLimiter(logging.GetComponentLogger("ratelimit"))

	// Protected routes (auth required)
	protected := router.Group("")
	protected.Use(middlewares.AuthMiddleware())
	protected.Use(rateLimiter.Middleware(middlewares.ProtectedEndpointLimit))
	{
		// Conversation management
		protected.POST("/otto/conversations", handler.Handler.CreateConversation)
		protected.GET("/otto/conversations", handler.Handler.ListConversations)
		protected.GET("/otto/conversations/:id", handler.Handler.GetConversation)
		protected.PUT("/otto/conversations/:id", handler.Handler.UpdateConversation)
		protected.POST("/otto/conversations/:id/archive", handler.Handler.ArchiveConversation)
		protected.POST("/otto/conversations/:id/bookmark", handler.Handler.BookmarkConversation)
		protected.DELETE("/otto/conversations/:id", handler.Handler.DeleteConversation)

		// Messages and chat
		protected.POST("/otto/conversations/:id/messages", handler.Handler.SendMessage)
		protected.GET("/otto/conversations/:id/messages", handler.Handler.GetMessages)
		protected.POST("/otto/messages/:id/feedback", handler.Handler.AddMessageFeedback)

		// Context enrichment
		protected.GET("/otto/conversations/:id/context", handler.Handler.GetConversationContext)
	}
}
