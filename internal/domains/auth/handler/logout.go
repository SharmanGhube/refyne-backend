package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"go.uber.org/zap"
)

// Logout handles user logout by blacklisting the current token
func (h *AuthHandlerImpl) Logout(c *gin.Context) {
	h.logger.Info("Logout request", zap.String("requestID", middlewares.GetRequestID(c)))

	// Get token from context (set by auth middleware)
	token, exists := middlewares.GetToken(c)
	if !exists || token == "" {
		h.logger.Warn("Logout attempt without token", zap.String("requestID", middlewares.GetRequestID(c)))
		middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "No token provided", nil)
		return
	}

	// Get user info for logging
	userID, _ := middlewares.GetUserID(c)

	// Call service to blacklist token
	if appErr := h.authService.Logout(c, token); appErr != nil {
		h.logger.Error("Logout failed",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.String("userID", userID),
			zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("User logged out successfully",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", userID))

	middlewares.RespondWithSuccess(c, http.StatusOK, "Logged out successfully", gin.H{
		"status": "logged_out",
	})
}

// LogoutAllDevices handles logging out from all devices by blacklisting all user tokens
func (h *AuthHandlerImpl) LogoutAllDevices(c *gin.Context) {
	h.logger.Info("Logout all devices request", zap.String("requestID", middlewares.GetRequestID(c)))

	// Get user ID from context (set by auth middleware)
	userID, exists := middlewares.GetUserID(c)
	if !exists || userID == "" {
		h.logger.Warn("Logout all devices attempt without authentication",
			zap.String("requestID", middlewares.GetRequestID(c)))
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		return
	}

	// Call service to blacklist all user tokens
	if appErr := h.authService.LogoutAllDevices(c, userID); appErr != nil {
		h.logger.Error("Logout all devices failed",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.String("userID", userID),
			zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("User logged out from all devices",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", userID))

	middlewares.RespondWithSuccess(c, http.StatusOK, "Logged out from all devices successfully", gin.H{
		"status": "logged_out_all_devices",
	})
}
