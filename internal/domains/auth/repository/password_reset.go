package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// PasswordResetToken represents a password reset token in the database
type PasswordResetToken struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	Token     string     `db:"token"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	IsValid   bool       `db:"is_valid"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type PasswordResetRepository interface {
	// Create a new password reset token
	CreateResetToken(c *gin.Context, userID, token string, expiresAt time.Time) *errors.AppError

	// Get reset token by token string
	GetResetToken(c *gin.Context, token string) (*PasswordResetToken, *errors.AppError)

	// Mark token as used
	MarkTokenAsUsed(c *gin.Context, tokenID string) *errors.AppError

	// Invalidate all tokens for a user
	InvalidateUserTokens(c *gin.Context, userID string) *errors.AppError

	// Delete expired tokens (cleanup)
	DeleteExpiredTokens(c *gin.Context) *errors.AppError
}

type PasswordResetRepositoryImpl struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

func NewPasswordResetRepository(db *sqlx.DB) PasswordResetRepository {
	return &PasswordResetRepositoryImpl{
		name:   "PasswordResetRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("PasswordResetRepository"),
	}
}

// CreateResetToken creates a new password reset token
func (r *PasswordResetRepositoryImpl) CreateResetToken(c *gin.Context, userID, token string, expiresAt time.Time) *errors.AppError {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(c.Request.Context(), query, userID, token, expiresAt)
	if err != nil {
		r.logger.Error("Failed to create password reset token",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return errors.NewAppError(
			c,
			"CREATE_RESET_TOKEN_FAILED",
			"Failed to create password reset token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"auth-repository",
		)
	}

	r.logger.Info("Password reset token created",
		zap.String("user_id", userID),
	)

	return nil
}

// GetResetToken retrieves a reset token by token string
func (r *PasswordResetRepositoryImpl) GetResetToken(c *gin.Context, token string) (*PasswordResetToken, *errors.AppError) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, is_valid, created_at, updated_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	var resetToken PasswordResetToken
	err := r.db.GetContext(c.Request.Context(), &resetToken, query, token)
	if err != nil {
		r.logger.Error("Failed to get password reset token",
			zap.Error(err),
			zap.String("token", token[:10]+"..."), // Log only first 10 chars for security
		)
		return nil, errors.NewAppError(
			c,
			"RESET_TOKEN_NOT_FOUND",
			"Password reset token not found",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"auth-repository",
		)
	}

	return &resetToken, nil
}

// MarkTokenAsUsed marks a token as used
func (r *PasswordResetRepositoryImpl) MarkTokenAsUsed(c *gin.Context, tokenID string) *errors.AppError {
	query := `
		UPDATE password_reset_tokens
		SET used_at = CURRENT_TIMESTAMP, is_valid = false
		WHERE id = $1
	`

	_, err := r.db.ExecContext(c.Request.Context(), query, tokenID)
	if err != nil {
		r.logger.Error("Failed to mark token as used",
			zap.Error(err),
			zap.String("token_id", tokenID),
		)
		return errors.NewAppError(
			c,
			"MARK_TOKEN_USED_FAILED",
			"Failed to update token status",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"auth-repository",
		)
	}

	r.logger.Info("Password reset token marked as used",
		zap.String("token_id", tokenID),
	)

	return nil
}

// InvalidateUserTokens invalidates all reset tokens for a user
func (r *PasswordResetRepositoryImpl) InvalidateUserTokens(c *gin.Context, userID string) *errors.AppError {
	query := `
		UPDATE password_reset_tokens
		SET is_valid = false
		WHERE user_id = $1 AND is_valid = true
	`

	_, err := r.db.ExecContext(c.Request.Context(), query, userID)
	if err != nil {
		r.logger.Error("Failed to invalidate user tokens",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return errors.NewAppError(
			c,
			"INVALIDATE_TOKENS_FAILED",
			"Failed to invalidate reset tokens",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"auth-repository",
		)
	}

	r.logger.Info("User password reset tokens invalidated",
		zap.String("user_id", userID),
	)

	return nil
}

// DeleteExpiredTokens deletes expired reset tokens (cleanup)
func (r *PasswordResetRepositoryImpl) DeleteExpiredTokens(c *gin.Context) *errors.AppError {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
	`

	result, err := r.db.ExecContext(c.Request.Context(), query)
	if err != nil {
		r.logger.Error("Failed to delete expired tokens",
			zap.Error(err),
		)
		return errors.NewAppError(
			c,
			"DELETE_EXPIRED_TOKENS_FAILED",
			"Failed to cleanup expired tokens",
			errors.ErrorTypeInternal,
			errors.SeverityLow,
			"auth-repository",
		)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("Expired password reset tokens deleted",
		zap.Int64("count", rowsAffected),
	)

	return nil
}
