package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// FailedAttempt represents a failed login attempt
type FailedAttempt struct {
	ID          string    `db:"id"`
	UserID      *string   `db:"user_id"`
	Email       string    `db:"email"`
	IPAddress   string    `db:"ip_address"`
	AttemptType string    `db:"attempt_type"`
	AttemptedAt time.Time `db:"attempted_at"`
	CreatedAt   time.Time `db:"created_at"`
}

// AccountLockout represents an account lockout
type AccountLockout struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	LockedUntil time.Time `db:"locked_until"`
	Reason      string    `db:"reason"`
	LockCount   int       `db:"lock_count"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// AccountSecurityRepository defines operations for account security
type AccountSecurityRepository interface {
	// Failed attempts
	RecordFailedAttempt(c *gin.Context, userID *string, email, ipAddress, attemptType string) *errors.AppError
	GetFailedAttemptsCount(c *gin.Context, email string, since time.Time) (int, *errors.AppError)
	ClearFailedAttempts(c *gin.Context, email string) *errors.AppError
	
	// Account lockouts
	LockAccount(c *gin.Context, userID, reason string, duration time.Duration) *errors.AppError
	IsAccountLocked(c *gin.Context, userID string) (bool, time.Time, *errors.AppError)
	UnlockAccount(c *gin.Context, userID string) *errors.AppError
	CleanupExpiredLockouts(c *gin.Context) *errors.AppError
}

type accountSecurityRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewAccountSecurityRepository creates a new account security repository
func NewAccountSecurityRepository(db *sqlx.DB) AccountSecurityRepository {
	return &accountSecurityRepository{
		db:     db,
		logger: logging.GetRepositoryLogger("AccountSecurityRepository"),
	}
}

// RecordFailedAttempt logs a failed login attempt
func (r *accountSecurityRepository) RecordFailedAttempt(
	c *gin.Context,
	userID *string,
	email, ipAddress, attemptType string,
) *errors.AppError {
	r.logger.Info("Recording failed attempt",
		zap.String("email", email),
		zap.String("ip", ipAddress),
		zap.String("type", attemptType),
	)
	
	query := `
		INSERT INTO failed_login_attempts (user_id, email, ip_address, attempt_type)
		VALUES ($1, $2, $3, $4)
	`
	
	_, err := r.db.ExecContext(c, query, userID, email, ipAddress, attemptType)
	if err != nil {
		r.logger.Error("Failed to record failed attempt", zap.Error(err))
		return errors.NewAppError(
			c,
			"FAILED_ATTEMPT_RECORD_ERROR",
			"Failed to record failed attempt",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"auth",
		)
	}
	
	return nil
}

// GetFailedAttemptsCount returns the count of failed attempts since a given time
func (r *accountSecurityRepository) GetFailedAttemptsCount(
	c *gin.Context,
	email string,
	since time.Time,
) (int, *errors.AppError) {
	query := `
		SELECT COUNT(*) 
		FROM failed_login_attempts 
		WHERE email = $1 AND attempted_at > $2
	`
	
	var count int
	err := r.db.GetContext(c, &count, query, email, since)
	if err != nil {
		r.logger.Error("Failed to get failed attempts count", zap.Error(err))
		return 0, errors.NewAppError(
			c,
			"FAILED_ATTEMPTS_COUNT_ERROR",
			"Failed to retrieve failed attempts count",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"auth",
		)
	}
	
	return count, nil
}

// ClearFailedAttempts removes all failed attempts for an email
func (r *accountSecurityRepository) ClearFailedAttempts(
	c *gin.Context,
	email string,
) *errors.AppError {
	query := `DELETE FROM failed_login_attempts WHERE email = $1`
	
	_, err := r.db.ExecContext(c, query, email)
	if err != nil {
		r.logger.Error("Failed to clear failed attempts", zap.Error(err))
		return errors.NewAppError(
			c,
			"CLEAR_FAILED_ATTEMPTS_ERROR",
			"Failed to clear failed attempts",
			errors.ErrorTypeInternal,
			errors.SeverityLow,
			"auth",
		)
	}
	
	r.logger.Info("Cleared failed attempts", zap.String("email", email))
	return nil
}

// LockAccount locks a user account for a specified duration
func (r *accountSecurityRepository) LockAccount(
	c *gin.Context,
	userID, reason string,
	duration time.Duration,
) *errors.AppError {
	r.logger.Info("Locking account",
		zap.String("user_id", userID),
		zap.String("reason", reason),
		zap.Duration("duration", duration),
	)
	
	lockedUntil := time.Now().Add(duration)
	
	query := `
		INSERT INTO account_lockouts (user_id, locked_until, reason, lock_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (user_id) DO UPDATE
		SET locked_until = $2,
		    reason = $3,
		    lock_count = account_lockouts.lock_count + 1,
		    updated_at = CURRENT_TIMESTAMP
	`
	
	_, err := r.db.ExecContext(c, query, userID, lockedUntil, reason)
	if err != nil {
		r.logger.Error("Failed to lock account", zap.Error(err))
		return errors.NewAppError(
			c,
			"ACCOUNT_LOCK_ERROR",
			"Failed to lock account",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"auth",
		)
	}
	
	return nil
}

// IsAccountLocked checks if an account is currently locked
func (r *accountSecurityRepository) IsAccountLocked(
	c *gin.Context,
	userID string,
) (bool, time.Time, *errors.AppError) {
	query := `
		SELECT locked_until 
		FROM account_lockouts 
		WHERE user_id = $1 AND locked_until > CURRENT_TIMESTAMP
	`
	
	var lockedUntil time.Time
	err := r.db.GetContext(c, &lockedUntil, query, userID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return false, time.Time{}, nil
		}
		r.logger.Error("Failed to check account lock status", zap.Error(err))
		return false, time.Time{}, errors.NewAppError(
			c,
			"LOCK_CHECK_ERROR",
			"Failed to check account lock status",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"auth",
		)
	}
	
	return true, lockedUntil, nil
}

// UnlockAccount removes the lockout for a user account
func (r *accountSecurityRepository) UnlockAccount(
	c *gin.Context,
	userID string,
) *errors.AppError {
	query := `DELETE FROM account_lockouts WHERE user_id = $1`
	
	_, err := r.db.ExecContext(c, query, userID)
	if err != nil {
		r.logger.Error("Failed to unlock account", zap.Error(err))
		return errors.NewAppError(
			c,
			"ACCOUNT_UNLOCK_ERROR",
			"Failed to unlock account",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"auth",
		)
	}
	
	r.logger.Info("Account unlocked", zap.String("user_id", userID))
	return nil
}

// CleanupExpiredLockouts removes expired lockouts
func (r *accountSecurityRepository) CleanupExpiredLockouts(c *gin.Context) *errors.AppError {
	query := `DELETE FROM account_lockouts WHERE locked_until < CURRENT_TIMESTAMP`
	
	result, err := r.db.ExecContext(c, query)
	if err != nil {
		r.logger.Error("Failed to cleanup expired lockouts", zap.Error(err))
		return errors.NewAppError(
			c,
			"CLEANUP_LOCKOUTS_ERROR",
			"Failed to cleanup expired lockouts",
			errors.ErrorTypeInternal,
			errors.SeverityLow,
			"auth",
		)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.logger.Info("Cleaned up expired lockouts", zap.Int64("count", rowsAffected))
	}
	
	return nil
}
