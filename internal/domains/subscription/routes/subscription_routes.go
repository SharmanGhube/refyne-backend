package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	handlerregistry "github.com/refynehq/refyne-backend/internal/shared/handlerRegistry"
	"github.com/refynehq/refyne-backend/pkg/logging"
)

// SetupSubscriptionRoutes configures all subscription-related routes
func SetupSubscriptionRoutes(router *gin.RouterGroup, registry *handlerregistry.HandlerRegistry) {
	subscriptionHandler := registry.Subscription.SubscriptionHandler
	rateLimiter := middlewares.NewInMemoryRateLimiter(logging.GetComponentLogger("ratelimit"))

	subscriptionGroup := router.Group("/subscription")
	{
		// Protected routes (require authentication)
		protected := subscriptionGroup.Group("")
		protected.Use(middlewares.AuthMiddleware())
		{
			// POST /api/subscription/checkout - Create checkout session
			// Rate limit: 10 requests per hour per user
			protected.POST("/checkout",
				rateLimiter.Middleware(middlewares.RateLimitRule{
					Requests: 10,
					Window:   time.Hour,
					KeyFunc: func(c *gin.Context) string {
						userID, _ := c.Get("user_id")
						return "subscription:checkout:" + userID.(string)
					},
				}),
				subscriptionHandler.CreateCheckout,
			)

			// GET /api/subscription/status - Get subscription status
			protected.GET("/status", subscriptionHandler.GetSubscriptionStatus)

			// POST /api/subscription/portal - Get customer portal URL
			protected.POST("/portal", subscriptionHandler.GetCustomerPortal)
		}
	}

	// Webhook route (public, but signature-verified)
	webhookGroup := router.Group("/webhooks")
	{
		// POST /api/webhooks/paddle - Handle Paddle webhooks
		// No auth middleware - signature verification in handler
		webhookGroup.POST("/paddle", subscriptionHandler.HandleWebhook)
	}
}
