package auth

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Password reset token validity duration (1 hour)
	ResetTokenValidity = 1 * time.Hour
	// Reset token length in bytes (will be 64 chars when hex encoded)
	ResetTokenLength = 32
)

// RequestPasswordReset initiates the password reset process
func (s *AuthServiceImpl) RequestPasswordReset(c *gin.Context, email string) *errors.AppError {
	s.logger.Info("Password reset requested",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("email", email),
	)

	// Check if user exists
	existingUser, err := s.coreUserRepo.GetUserByEmail(c, email)
	if err != nil {
		return err
	}
	if existingUser == nil {
		// Don't reveal if user exists or not for security
		s.logger.Warn("Password reset requested for non-existent user",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.String("email", email),
		)
		return nil // Return success to prevent email enumeration
	}

	// Check if user is active
	if !existingUser.IsActive {
		s.logger.Warn("Password reset requested for inactive user",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.String("email", email),
		)
		return errors.NewAppError(
			c,
			"USER_INACTIVE",
			"Account is not active",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	// Invalidate any existing reset tokens for this user
	if err := s.passwordResetRepo.InvalidateUserTokens(c, existingUser.ID); err != nil {
		return err
	}

	// Generate secure reset token
	token, genErr := generateSecureToken(ResetTokenLength)
	if genErr != nil {
		s.logger.Error("Failed to generate reset token",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(genErr),
		)
		return errors.NewAppError(
			c,
			"TOKEN_GENERATION_FAILED",
			"Failed to generate reset token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"auth",
		)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(ResetTokenValidity)

	// Store reset token
	if err := s.passwordResetRepo.CreateResetToken(c, existingUser.ID, token, expiresAt); err != nil {
		return err
	}

	// Send password reset email
	// Construct reset link using configured frontend URL
	resetLink := s.frontendURL + "/reset-password?token=" + token
	if s.emailService != nil {
		if emailErr := s.emailService.SendPasswordReset(email, token, resetLink); emailErr != nil {
			s.logger.Error("Failed to send password reset email",
				zap.String("requestID", middlewares.GetRequestID(c)),
				zap.String("email", email),
				zap.Error(emailErr),
			)
			// Don't fail the request if email sending fails
		}
	} else {
		s.logger.Warn("Email service not configured, reset link not sent via email")
	}

	s.logger.Info("Password reset token created",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("user_id", existingUser.ID),
	)

	return nil
}

// ValidateResetToken validates a password reset token
func (s *AuthServiceImpl) ValidateResetToken(c *gin.Context, token string) (*string, *errors.AppError) {
	s.logger.Info("Validating reset token",
		zap.String("requestID", middlewares.GetRequestID(c)),
	)

	// Get token from database
	resetToken, err := s.passwordResetRepo.GetResetToken(c, token)
	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !resetToken.IsValid {
		s.logger.Warn("Invalid reset token used",
			zap.String("requestID", middlewares.GetRequestID(c)),
		)
		return nil, errors.NewAppError(
			c,
			"INVALID_RESET_TOKEN",
			"Invalid or expired reset token",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	// Check if token has been used
	if resetToken.UsedAt != nil {
		s.logger.Warn("Used reset token attempted",
			zap.String("requestID", middlewares.GetRequestID(c)),
		)
		return nil, errors.NewAppError(
			c,
			"TOKEN_ALREADY_USED",
			"Reset token has already been used",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	// Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		s.logger.Warn("Expired reset token used",
			zap.String("requestID", middlewares.GetRequestID(c)),
		)
		return nil, errors.NewAppError(
			c,
			"TOKEN_EXPIRED",
			"Reset token has expired",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	s.logger.Info("Reset token validated successfully",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("user_id", resetToken.UserID),
	)

	return &resetToken.UserID, nil
}

// ResetPassword resets the user's password using a valid reset token
func (s *AuthServiceImpl) ResetPassword(c *gin.Context, token, newPassword string) *errors.AppError {
	s.logger.Info("Resetting password",
		zap.String("requestID", middlewares.GetRequestID(c)),
	)

	// Validate token and get user ID
	userID, err := s.ValidateResetToken(c, token)
	if err != nil {
		return err
	}

	// Get user
	user, err := s.coreUserRepo.GetUserByID(c, *userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.NewAppError(
			c,
			"USER_NOT_FOUND",
			"User not found",
			errors.ErrorTypeNotFound,
			errors.SeverityMedium,
			"auth",
		)
	}

	// Hash new password
	hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if hashErr != nil {
		s.logger.Error("Failed to hash password",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(hashErr),
		)
		return errors.NewAppError(
			c,
			"PASSWORD_HASH_FAILED",
			"Failed to reset password",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"auth",
		)
	}

	// Update password using repository method
	if err := s.coreUserRepo.UpdatePassword(c, *userID, string(hashedPassword)); err != nil {
		s.logger.Error("Failed to update password",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err),
		)
		return err
	}

	// Mark token as used
	resetToken, _ := s.passwordResetRepo.GetResetToken(c, token)
	if resetToken != nil {
		if err := s.passwordResetRepo.MarkTokenAsUsed(c, resetToken.ID); err != nil {
			// Log error but don't fail the operation
			s.logger.Warn("Failed to mark token as used",
				zap.String("requestID", middlewares.GetRequestID(c)),
				zap.Error(err),
			)
		}
	}

	// Invalidate all other reset tokens for this user
	if err := s.passwordResetRepo.InvalidateUserTokens(c, *userID); err != nil {
		// Log error but don't fail the operation
		s.logger.Warn("Failed to invalidate user tokens",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err),
		)
	}

	// TODO: Send password change confirmation email (Phase 1.3)

	s.logger.Info("Password reset successfully",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("user_id", *userID),
	)

	return nil
}

// Helper function to extract DB from repository (uses reflection-like approach)
func getDBFromRepo(repo interface{}) *sqlx.DB {
	type dbGetter interface {
		GetDB() *sqlx.DB
	}
	if getter, ok := repo.(dbGetter); ok {
		return getter.GetDB()
	}
	// Fallback - this will be handled by adding GetDB method to repository
	return nil
}

// generateSecureToken generates a cryptographically secure random token
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
