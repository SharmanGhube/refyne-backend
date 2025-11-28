package services

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// PaddleService defines the interface for interacting with Paddle API
// This abstraction allows us to swap between mock, sandbox, and production implementations
type PaddleService interface {
	// GenerateCheckoutURL creates a Paddle checkout session and returns the URL
	GenerateCheckoutURL(ctx *gin.Context, userID, userEmail, tier string) (string, *errors.AppError)

	// GetCustomerPortalURL generates a URL to Paddle's customer portal for subscription management
	GetCustomerPortalURL(ctx *gin.Context, customerID string) (string, *errors.AppError)

	// VerifyWebhookSignature validates that a webhook came from Paddle
	VerifyWebhookSignature(payload []byte, signature string) bool

	// GetMode returns the current payment mode (mock, sandbox, production)
	GetMode() string
}

// WebhookService defines the interface for processing Paddle webhook events
type WebhookService interface {
	// ProcessWebhook handles incoming Paddle webhook events
	ProcessWebhook(ctx *gin.Context, event *models.PaddleWebhookEvent) *errors.AppError

	// ProcessSubscriptionCreated handles subscription.created event
	ProcessSubscriptionCreated(ctx *gin.Context, data map[string]interface{}) *errors.AppError

	// ProcessSubscriptionUpdated handles subscription.updated event
	ProcessSubscriptionUpdated(ctx *gin.Context, data map[string]interface{}) *errors.AppError

	// ProcessSubscriptionCanceled handles subscription.canceled event
	ProcessSubscriptionCanceled(ctx *gin.Context, data map[string]interface{}) *errors.AppError

	// ProcessSubscriptionPastDue handles subscription.past_due event
	ProcessSubscriptionPastDue(ctx *gin.Context, data map[string]interface{}) *errors.AppError

	// ProcessSubscriptionPaused handles subscription.paused event
	ProcessSubscriptionPaused(ctx *gin.Context, data map[string]interface{}) *errors.AppError

	// ProcessSubscriptionResumed handles subscription.resumed event
	ProcessSubscriptionResumed(ctx *gin.Context, data map[string]interface{}) *errors.AppError

	// ProcessTransactionCompleted handles transaction.completed event
	ProcessTransactionCompleted(ctx *gin.Context, data map[string]interface{}) *errors.AppError

	// ProcessTransactionPaymentFailed handles transaction.payment_failed event
	ProcessTransactionPaymentFailed(ctx *gin.Context, data map[string]interface{}) *errors.AppError
}
