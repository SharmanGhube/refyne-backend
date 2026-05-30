package auth

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

const (
	// Verification token validity duration (24 hours)
	VerificationTokenValidity = 24 * time.Hour
	// Verification token length in bytes (will be 64 chars when hex encoded)
	VerificationTokenLength = 32
)

// SendVerificationEmail sends a verification email to the user
func (s *AuthServiceImpl) SendVerificationEmail(c *gin.Context, userID, email, username string) *errors.AppError {
	s.logger.Info("Sending verification email",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("user_id", userID),
		zap.String("email", email),
	)

	// Invalidate any existing verification tokens for this user
	if err := s.verificationRepo.InvalidateUserTokens(c, userID); err != nil {
		s.logger.Warn("Failed to invalidate existing verification tokens",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		// Don't fail the request, continue with new token generation
	}

	// Generate secure verification token
	token, genErr := generateSecureToken(VerificationTokenLength)
	if genErr != nil {
		s.logger.Error("Failed to generate verification token",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(genErr),
		)
		return errors.NewAppError(
			c,
			"TOKEN_GENERATION_FAILED",
			"Failed to generate verification token",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"auth",
		)
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(VerificationTokenValidity)

	// Store verification token
	if err := s.verificationRepo.CreateVerificationToken(c, userID, token, expiresAt); err != nil {
		return err
	}

	// Send verification email
	verificationLink := s.frontendURL + "/auth/verify-email?token=" + token
	if s.emailService != nil {
		if emailErr := s.emailService.SendVerification(email, username, verificationLink); emailErr != nil {
			s.logger.Error("Failed to send verification email",
				zap.String("requestID", middlewares.GetRequestID(c)),
				zap.String("email", email),
				zap.Error(emailErr),
			)
			// Don't fail the request if email sending fails
		} else {
			s.logger.Info("Verification email sent successfully",
				zap.String("requestID", middlewares.GetRequestID(c)),
				zap.String("user_id", userID),
			)
		}
	} else {
		s.logger.Warn("Email service not configured, verification email not sent",
			zap.String("user_id", userID),
		)
	}

	return nil
}

// VerifyAccount verifies a user's email using the verification token
func (s *AuthServiceImpl) VerifyAccount(c *gin.Context, token string) *errors.AppError {
	s.logger.Info("Verifying account",
		zap.String("requestID", middlewares.GetRequestID(c)),
	)

	// Get verification token from database
	vToken, err := s.verificationRepo.GetVerificationToken(c, token)
	if err != nil {
		return err
	}

	// Check if token is valid
	if !vToken.IsValid {
		s.logger.Warn("Verification token already used",
			zap.String("requestID", middlewares.GetRequestID(c)),
		)
		return errors.NewAppError(
			c,
			"VERIFICATION_TOKEN_USED",
			"This verification link has already been used",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	// Check if token has expired
	if time.Now().After(vToken.ExpiresAt) {
		s.logger.Warn("Verification token expired",
			zap.String("requestID", middlewares.GetRequestID(c)),
		)
		return errors.NewAppError(
			c,
			"VERIFICATION_TOKEN_EXPIRED",
			"This verification link has expired. Please request a new one",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	// Get user
	user, getUserErr := s.coreUserRepo.GetUserByID(c, vToken.UserID)
	if getUserErr != nil {
		return getUserErr
	}

	// Check if user is already verified
	if user.IsVerified {
		s.logger.Info("User already verified",
			zap.String("user_id", vToken.UserID),
		)
		return errors.NewAppError(
			c,
			"USER_ALREADY_VERIFIED",
			"Your account is already verified",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	// Mark token as verified
	if err := s.verificationRepo.MarkTokenAsVerified(c, token); err != nil {
		return err
	}

	// Update user as verified and active
	updateErr := s.coreUserRepo.VerifyUser(c, vToken.UserID)
	if updateErr != nil {
		s.logger.Error("Failed to verify user",
			zap.String("user_id", vToken.UserID),
			zap.Error(updateErr),
		)
		return updateErr
	}

	s.logger.Info("Account verified successfully",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("user_id", vToken.UserID),
	)

	return nil
}

// ResendVerificationEmail resends verification email to user
func (s *AuthServiceImpl) ResendVerificationEmail(c *gin.Context, email string) *errors.AppError {
	s.logger.Info("Resending verification email",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("email", email),
	)

	// Get user by email
	user, err := s.coreUserRepo.GetUserByEmail(c, email)
	if err != nil {
		// Don't reveal if user exists
		s.logger.Warn("Verification email resend requested for non-existent user",
			zap.String("email", email),
		)
		return nil // Return success to prevent email enumeration
	}

	// Check if user is already verified
	if user.IsVerified {
		s.logger.Info("Verification resend requested for already verified user",
			zap.String("email", email),
		)
		return nil // Return success to prevent email enumeration
	}

	// Send new verification email
	return s.SendVerificationEmail(c, user.ID, user.Email, user.Username)
}

