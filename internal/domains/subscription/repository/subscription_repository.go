package repository

import (
	"time"

	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// SubscriptionRepository handles database operations for subscription management
type SubscriptionRepository interface {
	// UpdateUserSubscription updates user's subscription fields
	UpdateUserSubscription(ctx *gin.Context, userID, tier, status string, expiresAt *time.Time, customerID, subscriptionID *string) *errors.AppError

	// GetUserByPaddleCustomerID retrieves user by Paddle customer ID
	GetUserByPaddleCustomerID(ctx *gin.Context, customerID string) (string, *errors.AppError)

	// GetUserByPaddleSubscriptionID retrieves user by Paddle subscription ID
	GetUserByPaddleSubscriptionID(ctx *gin.Context, subscriptionID string) (string, *errors.AppError)

	// MarkOnboardingComplete marks onboarding as completed for user
	MarkOnboardingComplete(ctx *gin.Context, userID string) *errors.AppError

	// GetUserSubscriptionStatus retrieves user's subscription details
	GetUserSubscriptionStatus(ctx *gin.Context, userID string) (tier, status string, expiresAt *time.Time, customerID, subscriptionID *string, err *errors.AppError)
}
