package repository

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	subscriptionErrors "github.com/refynehq/refyne-backend/internal/domains/subscription/errors"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type SubscriptionRepositoryImpl struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *sqlx.DB) SubscriptionRepository {
	return &SubscriptionRepositoryImpl{
		name:   "SubscriptionRepository",
		db:     db,
		logger: logging.GetComponentLogger("subscription_repository"),
	}
}

// UpdateUserSubscription updates user's subscription fields
func (r *SubscriptionRepositoryImpl) UpdateUserSubscription(
	ctx *gin.Context,
	userID, tier, status string,
	expiresAt *time.Time,
	customerID, subscriptionID *string,
) *errors.AppError {
	query := `
		UPDATE users 
		SET 
			subscription_tier = $1,
			subscription_status = $2,
			subscription_expires_at = $3,
			paddle_customer_id = COALESCE($4, paddle_customer_id),
			paddle_subscription_id = COALESCE($5, paddle_subscription_id),
			updated_at = NOW()
		WHERE id = $6 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		tier,
		status,
		expiresAt,
		customerID,
		subscriptionID,
		userID,
	)

	if err != nil {
		r.logger.Error("Failed to update user subscription",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return subscriptionErrors.NewDatabaseError(ctx, "update_user_subscription", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No user found to update subscription",
			zap.String("user_id", userID),
		)
		return subscriptionErrors.NewUserNotFoundError(ctx, userID)
	}

	r.logger.Info("User subscription updated",
		zap.String("user_id", userID),
		zap.String("tier", tier),
		zap.String("status", status),
	)

	return nil
}

// GetUserByPaddleCustomerID retrieves user ID by Paddle customer ID
func (r *SubscriptionRepositoryImpl) GetUserByPaddleCustomerID(ctx *gin.Context, customerID string) (string, *errors.AppError) {
	query := `
		SELECT id 
		FROM users 
		WHERE paddle_customer_id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var userID string
	err := r.db.GetContext(ctx, &userID, query, customerID)

	if err == sql.ErrNoRows {
		r.logger.Warn("No user found for Paddle customer ID",
			zap.String("customer_id", customerID),
		)
		return "", subscriptionErrors.NewCustomerNotFoundError(ctx, customerID)
	}

	if err != nil {
		r.logger.Error("Failed to get user by Paddle customer ID",
			zap.String("customer_id", customerID),
			zap.Error(err),
		)
		return "", subscriptionErrors.NewDatabaseError(ctx, "get_user_by_customer_id", err)
	}

	return userID, nil
}

// GetUserByPaddleSubscriptionID retrieves user ID by Paddle subscription ID
func (r *SubscriptionRepositoryImpl) GetUserByPaddleSubscriptionID(ctx *gin.Context, subscriptionID string) (string, *errors.AppError) {
	query := `
		SELECT id 
		FROM users 
		WHERE paddle_subscription_id = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var userID string
	err := r.db.GetContext(ctx, &userID, query, subscriptionID)

	if err == sql.ErrNoRows {
		r.logger.Warn("No user found for Paddle subscription ID",
			zap.String("subscription_id", subscriptionID),
		)
		return "", subscriptionErrors.NewSubscriptionNotFoundError(ctx, subscriptionID)
	}

	if err != nil {
		r.logger.Error("Failed to get user by Paddle subscription ID",
			zap.String("subscription_id", subscriptionID),
			zap.Error(err),
		)
		return "", subscriptionErrors.NewDatabaseError(ctx, "get_user_by_subscription_id", err)
	}

	return userID, nil
}

// MarkOnboardingComplete marks onboarding as completed for user
func (r *SubscriptionRepositoryImpl) MarkOnboardingComplete(ctx *gin.Context, userID string) *errors.AppError {
	query := `
		UPDATE users 
		SET onboarding_completed = TRUE, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to mark onboarding complete",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return subscriptionErrors.NewDatabaseError(ctx, "mark_onboarding_complete", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return subscriptionErrors.NewUserNotFoundError(ctx, userID)
	}

	r.logger.Info("Onboarding marked complete",
		zap.String("user_id", userID),
	)

	return nil
}

// GetUserSubscriptionStatus retrieves user's subscription details
func (r *SubscriptionRepositoryImpl) GetUserSubscriptionStatus(
	ctx *gin.Context,
	userID string,
) (tier, status string, expiresAt *time.Time, customerID, subscriptionID *string, err *errors.AppError) {
	query := `
		SELECT 
			subscription_tier,
			subscription_status,
			subscription_expires_at,
			paddle_customer_id,
			paddle_subscription_id
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var result struct {
		Tier           string     `db:"subscription_tier"`
		Status         string     `db:"subscription_status"`
		ExpiresAt      *time.Time `db:"subscription_expires_at"`
		CustomerID     *string    `db:"paddle_customer_id"`
		SubscriptionID *string    `db:"paddle_subscription_id"`
	}

	dbErr := r.db.GetContext(ctx, &result, query, userID)
	if dbErr == sql.ErrNoRows {
		return "", "", nil, nil, nil, subscriptionErrors.NewUserNotFoundError(ctx, userID)
	}

	if dbErr != nil {
		r.logger.Error("Failed to get user subscription status",
			zap.String("user_id", userID),
			zap.Error(dbErr),
		)
		return "", "", nil, nil, nil, subscriptionErrors.NewDatabaseError(ctx, "get_subscription_status", dbErr)
	}

	return result.Tier, result.Status, result.ExpiresAt, result.CustomerID, result.SubscriptionID, nil
}
