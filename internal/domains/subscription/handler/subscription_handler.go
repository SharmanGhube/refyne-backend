package handler

import (
	"github.com/gin-gonic/gin"
)

// SubscriptionHandler defines the interface for subscription HTTP handlers
type SubscriptionHandler interface {
	// CreateCheckout generates a Paddle checkout URL for a subscription tier
	CreateCheckout(c *gin.Context)

	// GetSubscriptionStatus returns the current user's subscription details
	GetSubscriptionStatus(c *gin.Context)

	// GetCustomerPortal generates a URL to Paddle's customer portal
	GetCustomerPortal(c *gin.Context)

	// HandleWebhook processes Paddle webhook events
	HandleWebhook(c *gin.Context)
}
