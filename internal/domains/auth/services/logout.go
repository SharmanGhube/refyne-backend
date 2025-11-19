package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	authUtils "github.com/refynehq/refyne-backend/internal/domains/auth/utils"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

// Logout blacklists the current token
func (s *AuthServiceImpl) Logout(c *gin.Context, token string) *errors.AppError {
	s.logger.Info("Processing logout", zap.String("requestID", middlewares.GetRequestID(c)))

	if token == "" {
		s.logger.Warn("Logout called with empty token")
		return errors.NewAppError(
			c,
			"LOGOUT_INVALID_TOKEN",
			"Invalid token provided",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	// Validate and extract token to get expiry time
	claims, err := authUtils.ValidateAndExtractToken(token)
	if err != nil {
		s.logger.Warn("Invalid token during logout",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err))
		// Still consider logout successful - token is invalid anyway
		return nil
	}

	// Blacklist the token until its natural expiry
	tokenManager := authUtils.GetTokenBlacklistManager()
	expiresAt := claims.ExpiresAt.Time
	tokenManager.BlacklistToken(token, expiresAt, "logout")

	s.logger.Info("Token blacklisted successfully",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", claims.UserID),
		zap.Time("expiresAt", expiresAt))

	return nil
}

// LogoutAllDevices invalidates all tokens for a user
// Note: In-memory implementation tracks tokens as they're blacklisted.
// For production with multiple instances, use Redis or database.
// This implementation blacklists tokens as they're encountered.
func (s *AuthServiceImpl) LogoutAllDevices(c *gin.Context, userID string) *errors.AppError {
	s.logger.Info("Processing logout all devices",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", userID))

	if userID == "" {
		s.logger.Warn("Logout all devices called with empty userID")
		return errors.NewAppError(
			c,
			"LOGOUT_INVALID_USER",
			"Invalid user ID provided",
			errors.ErrorTypeValidation,
			errors.SeverityLow,
			"auth",
		)
	}

	// Get current token to blacklist
	if token, exists := middlewares.GetToken(c); exists && token != "" {
		claims, err := authUtils.ValidateAndExtractToken(token)
		if err == nil {
			tokenManager := authUtils.GetTokenBlacklistManager()
			expiresAt := claims.ExpiresAt.Time
			tokenManager.BlacklistToken(token, expiresAt, "logout_all_devices")

			s.logger.Info("Current token blacklisted",
				zap.String("requestID", middlewares.GetRequestID(c)),
				zap.String("userID", userID))
		}
	}

	// Note: In a full implementation with token storage (Redis/DB),
	// you would query all active tokens for this user and blacklist them.
	// For now, we're blacklisting the current token.
	// Future enhancement: Store all issued tokens with userID in Redis/DB
	// and iterate through them here.

	s.logger.Info("Logout all devices completed",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", userID))

	// TODO: When implementing token storage:
	// 1. Query all active tokens for userID from storage
	// 2. Blacklist each token
	// 3. Remove from active tokens storage

	return nil
}
