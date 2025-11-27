package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"go.uber.org/zap"
)

type RegisterRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Username  string `json:"username" binding:"required"`

	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// Handler for user registration
func (h *AuthHandlerImpl) Register(c *gin.Context) {
	var req RegisterRequest

	// Bind and validate the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid registration request", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Processing registration request", zap.String("requestID", middlewares.GetRequestID(c)))

	// Call Auth Service to register the user
	if appErr := h.authService.RegisterUser(c, req.FirstName, req.LastName, req.Username, req.Email, req.Password); appErr != nil {
		h.logger.Error("Registration failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("User registered successfully", zap.String("requestID", middlewares.GetRequestID(c)))

	// Respond with success - user needs to verify email
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. Please check your email to verify your account.",
	})
}

type OTPRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Handler for OTP request
func (h *AuthHandlerImpl) RequestOTP(c *gin.Context) {
	var req OTPRequest

	// Bind and validate the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid OTP request", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Processing OTP request", zap.String("requestID", middlewares.GetRequestID(c)))

	// Call Auth Service to generate OTP (sent via email)
	appErr := h.authService.RequestOTP(c, req.Email, req.Password)
	if appErr != nil {
		h.logger.Error("OTP request failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("OTP generated successfully", zap.String("requestID", middlewares.GetRequestID(c)))

	// Production-ready response (OTP sent via email only)
	// Note: In production, OTP should NEVER be included in the response
	// Users receive OTP via email only for security
	c.JSON(http.StatusOK, gin.H{
		"message":    "OTP sent successfully to your email",
		"expires_in": 900, // 15 minutes in seconds
		"RequestID":  middlewares.GetRequestID(c),
	})
}

type OTPVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

// Handler for OTP verification and login
func (h *AuthHandlerImpl) VerifyOTP(c *gin.Context) {
	var req OTPVerifyRequest

	// Bind and validate the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid OTP verification request", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Processing OTP verification", zap.String("requestID", middlewares.GetRequestID(c)))

	// Call Auth Service to verify OTP and login
	user, tokenPair, appErr := h.authService.VerifyOTPAndLogin(c, req.Email, req.OTP)
	if appErr != nil {
		h.logger.Error("OTP verification failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("OTP verified and user logged in successfully", zap.String("requestID", middlewares.GetRequestID(c)))

	// Prepare user response (exclude sensitive data)
	userResponse := gin.H{
		"id":          user.ID,
		"email":       user.Email,
		"username":    user.Username,
		"first_name":  user.FirstName,
		"last_name":   user.LastName,
		"status":      user.Status,
		"is_active":   user.IsActive,
		"is_verified": user.IsVerified,
	}

	// Respond with the JWT tokens
	c.JSON(http.StatusOK, gin.H{
		"message":   "Login successful",
		"user":      userResponse,
		"TokenPair": tokenPair,
		"RequestID": middlewares.GetRequestID(c),
	})
}

// RefreshToken handles token refresh requests
func (h *AuthHandlerImpl) RefreshToken(c *gin.Context) {
	h.logger.Info("Refresh token request", zap.String("requestID", middlewares.GetRequestID(c)))

	// Request structure
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	// Bind request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid refresh token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Refresh token
	tokenPair, appErr := h.authService.RefreshToken(c, req.RefreshToken)
	if appErr != nil {
		h.logger.Error("Token refresh failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Token refreshed successfully",
		"TokenPair": tokenPair,
		"RequestID": middlewares.GetRequestID(c),
	})

	h.logger.Info("Token refresh successful", zap.String("requestID", middlewares.GetRequestID(c)))
}

// VerifyAccount handles account verification requests
func (h *AuthHandlerImpl) VerifyAccount(c *gin.Context) {
	h.logger.Info("Account verification request", zap.String("requestID", middlewares.GetRequestID(c)))

	// Request structure
	var req struct {
		Token string `json:"token" binding:"required"`
	}

	// Bind request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid verification request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Verify account
	if appErr := h.authService.VerifyAccount(c, req.Token); appErr != nil {
		h.logger.Error("Account verification failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	// Success response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Account verified successfully",
		"status":    "verified",
		"RequestID": middlewares.GetRequestID(c),
	})

	h.logger.Info("Account verification successful", zap.String("requestID", middlewares.GetRequestID(c)))
}

// ResendVerification handles resending verification email
func (h *AuthHandlerImpl) ResendVerification(c *gin.Context) {
	h.logger.Info("Resend verification request", zap.String("requestID", middlewares.GetRequestID(c)))

	// Request structure
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	// Bind request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid resend verification request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Resend verification email (always return success to prevent email enumeration)
	if appErr := h.authService.ResendVerificationEmail(c, req.Email); appErr != nil {
		// Log error but don't expose to client
		h.logger.Error("Resend verification failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
	}

	// Success response (always the same regardless of whether email exists)
	c.JSON(http.StatusOK, gin.H{
		"message":   "If an account exists with that email, a verification email has been sent.",
		"RequestID": middlewares.GetRequestID(c),
	})

	h.logger.Info("Resend verification request processed", zap.String("requestID", middlewares.GetRequestID(c)))
}
