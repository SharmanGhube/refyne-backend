package errors

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// Error codes
const (
	CodeDatabaseError           = "SUBSCRIPTION_DATABASE_ERROR"
	CodeUserNotFound            = "SUBSCRIPTION_USER_NOT_FOUND"
	CodeCustomerNotFound        = "SUBSCRIPTION_CUSTOMER_NOT_FOUND"
	CodeSubscriptionNotFound    = "SUBSCRIPTION_NOT_FOUND"
	CodeInvalidWebhook          = "SUBSCRIPTION_INVALID_WEBHOOK"
	CodeWebhookProcessing       = "SUBSCRIPTION_WEBHOOK_PROCESSING_ERROR"
	CodeCheckoutCreationFailed  = "SUBSCRIPTION_CHECKOUT_FAILED"
	CodePortalCreationFailed    = "SUBSCRIPTION_PORTAL_FAILED"
	CodeInvalidSubscriptionTier = "SUBSCRIPTION_INVALID_TIER"
)

// Repository errors
func NewDatabaseError(c *gin.Context, operation string, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeDatabaseError,
		"A database error occurred",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"subscription",
	).WithContext("operation", operation).WithContext("error", err.Error())
}

func NewUserNotFoundError(c *gin.Context, userID string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserNotFound,
		"User not found",
		errors.ErrorTypeNotFound,
		errors.SeverityLow,
		"subscription",
	).WithContext("user_id", userID)
}

func NewCustomerNotFoundError(c *gin.Context, customerID string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeCustomerNotFound,
		"Customer not found",
		errors.ErrorTypeNotFound,
		errors.SeverityLow,
		"subscription",
	).WithContext("customer_id", customerID)
}

func NewSubscriptionNotFoundError(c *gin.Context, subscriptionID string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeSubscriptionNotFound,
		"Subscription not found",
		errors.ErrorTypeNotFound,
		errors.SeverityLow,
		"subscription",
	).WithContext("subscription_id", subscriptionID)
}

// Service errors
func NewInvalidWebhookError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidWebhook,
		message,
		errors.ErrorTypeValidation,
		errors.SeverityMedium,
		"subscription",
	)
}

func NewWebhookProcessingError(c *gin.Context, eventType string, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeWebhookProcessing,
		"Failed to process webhook event",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"subscription",
	).WithContext("event_type", eventType).WithContext("error", err.Error())
}

func NewCheckoutCreationError(c *gin.Context, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeCheckoutCreationFailed,
		"Failed to create checkout session",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"subscription",
	).WithContext("error", err.Error())
}

func NewPortalCreationError(c *gin.Context, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodePortalCreationFailed,
		"Failed to create customer portal",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"subscription",
	).WithContext("error", err.Error())
}

func NewInvalidSubscriptionTierError(c *gin.Context, tier string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidSubscriptionTier,
		"Invalid subscription tier",
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"subscription",
	).WithContext("tier", tier)
}
