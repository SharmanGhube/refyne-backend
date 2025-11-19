package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"go.uber.org/zap"
)

// ForgotPasswordRequest represents the forgot password request body
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the reset password request body
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ValidateResetTokenRequest represents the validate token request body
type ValidateResetTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// ForgotPassword handles password reset request
func (h *AuthHandlerImpl) ForgotPassword(c *gin.Context) {
	h.logger.Info("Forgot password request received", zap.String("requestID", middlewares.GetRequestID(c)))

	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body. Email is required.",
		})
		return
	}

	// Request password reset
	if err := h.authService.RequestPasswordReset(c, req.Email); err != nil {
		h.logger.Error("Failed to request password reset",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err),
		)
		c.JSON(err.HTTPStatus, err.ClientResponse())
		return
	}

	// Always return success to prevent email enumeration
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset with token
func (h *AuthHandlerImpl) ResetPassword(c *gin.Context) {
	h.logger.Info("Reset password request received", zap.String("requestID", middlewares.GetRequestID(c)))

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body. Token and new password are required.",
		})
		return
	}

	// Reset password
	if err := h.authService.ResetPassword(c, req.Token, req.NewPassword); err != nil {
		h.logger.Error("Failed to reset password",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err),
		)
		c.JSON(err.HTTPStatus, err.ClientResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password has been reset successfully",
	})
}

// ValidateResetToken validates a password reset token
func (h *AuthHandlerImpl) ValidateResetToken(c *gin.Context) {
	h.logger.Info("Validate reset token request received", zap.String("requestID", middlewares.GetRequestID(c)))

	var req ValidateResetTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request body",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body. Token is required.",
		})
		return
	}

	// Validate token
	userID, err := h.authService.ValidateResetToken(c, req.Token)
	if err != nil {
		h.logger.Error("Failed to validate reset token",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.Error(err),
		)
		c.JSON(err.HTTPStatus, err.ClientResponse())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token is valid",
		"data": gin.H{
			"user_id": *userID,
		},
	})
}
