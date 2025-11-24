package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// VerificationToken represents an email verification token
type VerificationToken struct {
	ID         string     `db:"id"`
	UserID     string     `db:"user_id"`
	Token      string     `db:"token"`
	ExpiresAt  time.Time  `db:"expires_at"`
	IsValid    bool       `db:"is_valid"`
	VerifiedAt *time.Time `db:"verified_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

// VerificationRepository defines operations for verification tokens
type VerificationRepository interface {
	CreateVerificationToken(c *gin.Context, userID, token string, expiresAt time.Time) *errors.AppError
	GetVerificationToken(c *gin.Context, token string) (*VerificationToken, *errors.AppError)
	MarkTokenAsVerified(c *gin.Context, token string) *errors.AppError
	InvalidateUserTokens(c *gin.Context, userID string) *errors.AppError
	DeleteExpiredTokens(c *gin.Context) *errors.AppError
}

// verificationRepository implements VerificationRepository
type verificationRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewVerificationRepository creates a new verification repository
func NewVerificationRepository(db *sqlx.DB) VerificationRepository {
	return &verificationRepository{
		db:     db,
		logger: logging.GetRepositoryLogger("VerificationRepository"),
	}
}

// CreateVerificationToken creates a new verification token
func (r *verificationRepository) CreateVerificationToken(
	c *gin.Context,
	userID, token string,
	expiresAt time.Time,
) *errors.AppError {
	r.logger.Info("Creating verification token",
		zap.String("user_id", userID),
	)

	query := `
		INSERT INTO verification_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(c, query, userID, token, expiresAt)
	if err != nil {
		r.logger.Error("Failed to create verification token",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return errors.NewAppError(
			c,
			"VERIFICATION_TOKEN_CREATION_FAILED",
			"Failed to create verification token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"auth",
		)
	}

	r.logger.Info("Verification token created successfully",
		zap.String("user_id", userID),
	)
	return nil
}

// GetVerificationToken retrieves a verification token
func (r *verificationRepository) GetVerificationToken(
	c *gin.Context,
	token string,
) (*VerificationToken, *errors.AppError) {
	r.logger.Info("Getting verification token")

	query := `
		SELECT id, user_id, token, expires_at, is_valid, verified_at, created_at, updated_at
		FROM verification_tokens
		WHERE token = $1
	`

	var vToken VerificationToken
	err := r.db.GetContext(c, &vToken, query, token)
	if err != nil {
		r.logger.Error("Failed to get verification token", zap.Error(err))
		return nil, errors.NewAppError(
			c,
			"VERIFICATION_TOKEN_NOT_FOUND",
			"Verification token not found",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"auth",
		)
	}

	r.logger.Info("Verification token retrieved successfully")
	return &vToken, nil
}

// MarkTokenAsVerified marks a token as verified
func (r *verificationRepository) MarkTokenAsVerified(
	c *gin.Context,
	token string,
) *errors.AppError {
	r.logger.Info("Marking token as verified")

	query := `
		UPDATE verification_tokens
		SET is_valid = false, verified_at = CURRENT_TIMESTAMP
		WHERE token = $1 AND is_valid = true
	`

	result, err := r.db.ExecContext(c, query, token)
	if err != nil {
		r.logger.Error("Failed to mark token as verified", zap.Error(err))
		return errors.NewAppError(
			c,
			"TOKEN_VERIFICATION_FAILED",
			"Failed to mark token as verified",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"auth",
		)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewAppError(
			c,
			"VERIFICATION_TOKEN_INVALID",
			"Token is already used or invalid",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	r.logger.Info("Token marked as verified successfully")
	return nil
}

// InvalidateUserTokens invalidates all verification tokens for a user
func (r *verificationRepository) InvalidateUserTokens(
	c *gin.Context,
	userID string,
) *errors.AppError {
	r.logger.Info("Invalidating user verification tokens",
		zap.String("user_id", userID),
	)

	query := `
		UPDATE verification_tokens
		SET is_valid = false
		WHERE user_id = $1 AND is_valid = true
	`

	_, err := r.db.ExecContext(c, query, userID)
	if err != nil {
		r.logger.Error("Failed to invalidate user tokens",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return errors.NewAppError(
			c,
			"TOKEN_INVALIDATION_FAILED",
			"Failed to invalidate verification tokens",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"auth",
		)
	}

	r.logger.Info("User verification tokens invalidated")
	return nil
}

// DeleteExpiredTokens removes expired verification tokens
func (r *verificationRepository) DeleteExpiredTokens(c *gin.Context) *errors.AppError {
	r.logger.Info("Deleting expired verification tokens")

	query := `
		DELETE FROM verification_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
	`

	result, err := r.db.ExecContext(c, query)
	if err != nil {
		r.logger.Error("Failed to delete expired tokens", zap.Error(err))
		return errors.NewAppError(
			c,
			"TOKEN_DELETION_FAILED",
			"Failed to delete expired tokens",
			errors.ErrorTypeInternal,
			errors.SeverityLow,
			"auth",
		)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("Expired verification tokens deleted",
		zap.Int64("count", rowsAffected),
	)
	return nil
}
