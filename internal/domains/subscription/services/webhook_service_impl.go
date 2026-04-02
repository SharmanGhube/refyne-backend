package services

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/models"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/repository"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

type WebhookServiceImpl struct {
	name       string
	repository repository.SubscriptionRepository
	logger     *zap.Logger
}

// NewWebhookService creates a new webhook service
func NewWebhookService(repo repository.SubscriptionRepository, logger *zap.Logger) WebhookService {
	return &WebhookServiceImpl{
		name:       "WebhookService",
		repository: repo,
		logger:     logger,
	}
}

// ProcessWebhook routes the webhook event to the appropriate handler
func (w *WebhookServiceImpl) ProcessWebhook(ctx *gin.Context, event *models.PaddleWebhookEvent) *errors.AppError {
	w.logger.Info("Processing webhook event",
		zap.String("event_id", event.EventID),
		zap.String("event_type", event.EventType),
	)

	switch event.EventType {
	case "subscription.created", "subscription.activated":
		return w.ProcessSubscriptionCreated(ctx, event.Data)

	case "subscription.updated":
		return w.ProcessSubscriptionUpdated(ctx, event.Data)

	case "subscription.canceled":
		return w.ProcessSubscriptionCanceled(ctx, event.Data)

	case "subscription.past_due":
		return w.ProcessSubscriptionPastDue(ctx, event.Data)

	case "subscription.paused":
		return w.ProcessSubscriptionPaused(ctx, event.Data)

	case "subscription.resumed":
		return w.ProcessSubscriptionResumed(ctx, event.Data)

	case "transaction.completed":
		return w.ProcessTransactionCompleted(ctx, event.Data)

	case "transaction.payment_failed":
		return w.ProcessTransactionPaymentFailed(ctx, event.Data)

	default:
		w.logger.Warn("Unhandled webhook event type",
			zap.String("event_type", event.EventType),
		)
		// Don't return error for unhandled events - just log
		return nil
	}
}

// ProcessSubscriptionCreated handles subscription creation
func (w *WebhookServiceImpl) ProcessSubscriptionCreated(ctx *gin.Context, data map[string]interface{}) *errors.AppError {
	// Extract subscription data
	subscriptionID := getStringFromMap(data, "id")
	customerID := getStringFromMap(data, "customer_id")
	status := getStringFromMap(data, "status")

	// Extract custom data containing user ID
	customData, _ := data["custom_data"].(map[string]interface{})
	userID := getStringFromMap(customData, "user_id")

	if userID == "" {
		// Try to get user by customer ID
		var appErr *errors.AppError
		userID, appErr = w.repository.GetUserByPaddleCustomerID(ctx, customerID)
		if appErr != nil {
			w.logger.Error("Failed to find user for subscription",
				zap.String("customer_id", customerID),
			)
			return appErr
		}
	}

	// Extract product to determine tier (always "pro" now)
	items, _ := data["items"].([]interface{})
	tier := "pro"
	if len(items) > 0 {
		item, _ := items[0].(map[string]interface{})
		productID := getStringFromMap(item, "product_id")
		tier = w.mapProductIDToTier(productID)
	}

	// Extract billing period
	var expiresAt *time.Time
	if billingPeriod, ok := data["current_billing_period"].(map[string]interface{}); ok {
		if endsAtStr, ok := billingPeriod["ends_at"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, endsAtStr); err == nil {
				expiresAt = &parsed
			}
		}
	}

	// Map Paddle status to our status
	mappedStatus := w.mapPaddleStatus(status)

	// Update user subscription
	custID := &customerID
	subID := &subscriptionID
	appErr := w.repository.UpdateUserSubscription(ctx, userID, tier, mappedStatus, expiresAt, custID, subID)
	if appErr != nil {
		return appErr
	}

	w.logger.Info("Subscription created successfully",
		zap.String("user_id", userID),
		zap.String("tier", tier),
		zap.String("status", mappedStatus),
	)

	return nil
}

// ProcessSubscriptionUpdated handles subscription updates
func (w *WebhookServiceImpl) ProcessSubscriptionUpdated(ctx *gin.Context, data map[string]interface{}) *errors.AppError {
	subscriptionID := getStringFromMap(data, "id")
	status := getStringFromMap(data, "status")

	// Get user by subscription ID
	userID, appErr := w.repository.GetUserByPaddleSubscriptionID(ctx, subscriptionID)
	if appErr != nil {
		return appErr
	}

	// Extract new tier if changed (always "pro" now)
	items, _ := data["items"].([]interface{})
	tier := "pro"
	if len(items) > 0 {
		item, _ := items[0].(map[string]interface{})
		productID := getStringFromMap(item, "product_id")
		tier = w.mapProductIDToTier(productID)
	}

	// Extract new billing period
	var expiresAt *time.Time
	if billingPeriod, ok := data["current_billing_period"].(map[string]interface{}); ok {
		if endsAtStr, ok := billingPeriod["ends_at"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, endsAtStr); err == nil {
				expiresAt = &parsed
			}
		}
	}

	mappedStatus := w.mapPaddleStatus(status)

	// Update subscription
	appErr = w.repository.UpdateUserSubscription(ctx, userID, tier, mappedStatus, expiresAt, nil, nil)
	if appErr != nil {
		return appErr
	}

	w.logger.Info("Subscription updated",
		zap.String("user_id", userID),
		zap.String("tier", tier),
		zap.String("status", mappedStatus),
	)

	return nil
}

// ProcessSubscriptionCanceled handles subscription cancellation
func (w *WebhookServiceImpl) ProcessSubscriptionCanceled(ctx *gin.Context, data map[string]interface{}) *errors.AppError {
	subscriptionID := getStringFromMap(data, "id")

	userID, appErr := w.repository.GetUserByPaddleSubscriptionID(ctx, subscriptionID)
	if appErr != nil {
		return appErr
	}

	// Get current tier and set status to cancelled
	tier, _, expiresAt, _, _, appErr := w.repository.GetUserSubscriptionStatus(ctx, userID)
	if appErr != nil {
		return appErr
	}

	appErr = w.repository.UpdateUserSubscription(ctx, userID, tier, "cancelled", expiresAt, nil, nil)
	if appErr != nil {
		return appErr
	}

	w.logger.Info("Subscription canceled",
		zap.String("user_id", userID),
	)

	return nil
}

// ProcessSubscriptionPastDue handles past due subscriptions
func (w *WebhookServiceImpl) ProcessSubscriptionPastDue(ctx *gin.Context, data map[string]interface{}) *errors.AppError {
	subscriptionID := getStringFromMap(data, "id")

	userID, appErr := w.repository.GetUserByPaddleSubscriptionID(ctx, subscriptionID)
	if appErr != nil {
		return appErr
	}

	tier, _, expiresAt, _, _, appErr := w.repository.GetUserSubscriptionStatus(ctx, userID)
	if appErr != nil {
		return appErr
	}

	appErr = w.repository.UpdateUserSubscription(ctx, userID, tier, "past_due", expiresAt, nil, nil)
	if appErr != nil {
		return appErr
	}

	w.logger.Warn("Subscription past due",
		zap.String("user_id", userID),
	)

	return nil
}

// ProcessSubscriptionPaused handles subscription pause
func (w *WebhookServiceImpl) ProcessSubscriptionPaused(ctx *gin.Context, data map[string]interface{}) *errors.AppError {
	subscriptionID := getStringFromMap(data, "id")

	userID, appErr := w.repository.GetUserByPaddleSubscriptionID(ctx, subscriptionID)
	if appErr != nil {
		return appErr
	}

	tier, _, expiresAt, _, _, appErr := w.repository.GetUserSubscriptionStatus(ctx, userID)
	if appErr != nil {
		return appErr
	}

	appErr = w.repository.UpdateUserSubscription(ctx, userID, tier, "paused", expiresAt, nil, nil)
	if appErr != nil {
		return appErr
	}

	w.logger.Info("Subscription paused",
		zap.String("user_id", userID),
	)

	return nil
}

// ProcessSubscriptionResumed handles subscription resume
func (w *WebhookServiceImpl) ProcessSubscriptionResumed(ctx *gin.Context, data map[string]interface{}) *errors.AppError {
	subscriptionID := getStringFromMap(data, "id")

	userID, appErr := w.repository.GetUserByPaddleSubscriptionID(ctx, subscriptionID)
	if appErr != nil {
		return appErr
	}

	tier, _, _, _, _, appErr := w.repository.GetUserSubscriptionStatus(ctx, userID)
	if appErr != nil {
		return appErr
	}

	// Extract new billing period
	var expiresAt *time.Time
	if billingPeriod, ok := data["current_billing_period"].(map[string]interface{}); ok {
		if endsAtStr, ok := billingPeriod["ends_at"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, endsAtStr); err == nil {
				expiresAt = &parsed
			}
		}
	}

	appErr = w.repository.UpdateUserSubscription(ctx, userID, tier, "active", expiresAt, nil, nil)
	if appErr != nil {
		return appErr
	}

	w.logger.Info("Subscription resumed",
		zap.String("user_id", userID),
	)

	return nil
}

// ProcessTransactionCompleted handles successful transactions
func (w *WebhookServiceImpl) ProcessTransactionCompleted(ctx *gin.Context, data map[string]interface{}) *errors.AppError {
	w.logger.Info("Transaction completed",
		zap.String("transaction_id", getStringFromMap(data, "id")),
	)
	// Transaction completion is mainly for logging
	// Subscription status updates come from subscription events
	return nil
}

// ProcessTransactionPaymentFailed handles failed payments
func (w *WebhookServiceImpl) ProcessTransactionPaymentFailed(ctx *gin.Context, data map[string]interface{}) *errors.AppError {
	w.logger.Warn("Transaction payment failed",
		zap.String("transaction_id", getStringFromMap(data, "id")),
	)
	// Payment failure handling is done via subscription.past_due event
	return nil
}

// Helper functions

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (w *WebhookServiceImpl) mapProductIDToTier(productID string) string {
	// All subscriptions are "pro" tier now
	// Previously different product IDs mapped to different tiers,
	// but now we only support the Pro subscription tier
	if productID != "" {
		w.logger.Debug("Mapping product to pro tier",
			zap.String("product_id", productID),
		)
	}
	return "pro"
}

func (w *WebhookServiceImpl) mapPaddleStatus(paddleStatus string) string {
	statusMap := map[string]string{
		"active":   "active",
		"canceled": "cancelled",
		"past_due": "past_due",
		"paused":   "paused",
		"trialing": "trialing",
	}

	if status, ok := statusMap[paddleStatus]; ok {
		return status
	}

	return "inactive"
}

// MarshalJSON helper for debugging
func (w *WebhookServiceImpl) debugPrintData(data map[string]interface{}) {
	if bytes, err := json.MarshalIndent(data, "", "  "); err == nil {
		w.logger.Debug("Webhook data", zap.String("data", string(bytes)))
	}
}
