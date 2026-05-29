package auth

import (
	"net/http"
	"time"

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
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", map[string]interface{}{
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

	// Get the newly created user by email to return it
	user, appErr := h.authService.GetUserByEmail(c, req.Email)
	if appErr != nil {
		h.logger.Error("Failed to fetch registered user", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("User registered successfully", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", user.ID))

	// Build response data with complete user information
	responseData := gin.H{
		"user_id":              user.ID,
		"email":                user.Email,
		"first_name":           user.FirstName,
		"last_name":            user.LastName,
		"username":             user.Username,
		"is_verified":          user.IsVerified,
		"is_active":            user.IsActive,
		"status":               user.Status,
		"onboarding_completed": user.OnboardingCompleted,
		"subscription_status":  user.SubscriptionStatus,
		"subscription_tier":    user.SubscriptionTier,
		"created_at":           user.CreatedAt,
		"message":              "User registered successfully. Please check your email to verify your account.",
	}

	// Send response using standardized success envelope
	middlewares.RespondWithSuccess(c, http.StatusCreated, "User registered successfully", responseData)
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
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", map[string]interface{}{
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
	responseData := gin.H{
		"expires_in": 300, // 5 minutes in seconds (OTP validity period)
		"message":    "OTP sent successfully to your email",
	}
	middlewares.RespondWithSuccess(c, http.StatusOK, "OTP sent successfully", responseData)
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
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", map[string]interface{}{
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

	// Set refresh token as httpOnly cookie (not accessible via JS)
	middlewares.SetRefreshTokenCookie(c, tokenPair.RefreshToken, 7*24*time.Hour)

	// Respond with access token only — refresh token lives in the cookie
	responseData := gin.H{
		"user":         userResponse,
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Login successful", responseData)
}

// LoginWithPassword handles password-based login
func (h *AuthHandlerImpl) LoginWithPassword(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	// Bind and validate the request body
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(err))
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Processing password login request", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", req.Email))

	// Call Auth Service to verify credentials and login
	user, tokenPair, appErr := h.authService.LoginUser(c, req.Email, req.Password)
	if appErr != nil {
		h.logger.Error("Password login failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("User logged in successfully with password", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", user.ID))

	// Prepare user response (exclude sensitive data)
	userResponse := gin.H{
		"user_id":              user.ID,
		"email":                user.Email,
		"username":             user.Username,
		"first_name":           user.FirstName,
		"last_name":            user.LastName,
		"status":               user.Status,
		"is_active":            user.IsActive,
		"is_verified":          user.IsVerified,
		"onboarding_completed": user.OnboardingCompleted,
		"created_at":           user.CreatedAt,
	}

	// Set refresh token as httpOnly cookie (not accessible via JS)
	middlewares.SetRefreshTokenCookie(c, tokenPair.RefreshToken, 7*24*time.Hour)

	// Respond with access token only — refresh token lives in the cookie
	responseData := gin.H{
		"user":         userResponse,
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Login successful", responseData)
}

// RefreshToken handles token refresh requests.
// It reads the refresh token from the httpOnly cookie first, falling back to
// the JSON body for backward compatibility with older clients.
func (h *AuthHandlerImpl) RefreshToken(c *gin.Context) {
	h.logger.Info("Refresh token request", zap.String("requestID", middlewares.GetRequestID(c)))

	// 1. Try to read refresh token from httpOnly cookie (preferred)
	refreshToken, fromCookie := middlewares.GetRefreshTokenFromCookie(c)

	// 2. Fallback: read from JSON body (backward compat for older clients)
	if !fromCookie {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&req); err == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		h.logger.Warn("No refresh token found in cookie or body", zap.String("requestID", middlewares.GetRequestID(c)))
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Refresh token is required", nil)
		return
	}

	// Refresh the token pair
	tokenPair, appErr := h.authService.RefreshToken(c, refreshToken)
	if appErr != nil {
		h.logger.Error("Token refresh failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	// Rotate: set new refresh token in httpOnly cookie
	middlewares.SetRefreshTokenCookie(c, tokenPair.RefreshToken, 7*24*time.Hour)

	// Return new access token only
	responseData := gin.H{
		"access_token": tokenPair.AccessToken,
		"expires_in":   tokenPair.ExpiresIn,
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Token refreshed successfully", responseData)

	h.logger.Info("Token refresh successful", zap.String("requestID", middlewares.GetRequestID(c)))
}

// VerifyAccount handles account verification requests
func (h *AuthHandlerImpl) VerifyAccount(c *gin.Context) {
	h.logger.Info("Account verification request", zap.String("requestID", middlewares.GetRequestID(c)))

	// Get token from query param (GET) or JSON body (POST)
	var token string

	// Try query parameter first (for GET requests)
	token = c.Query("token")

	// If not in query, try JSON body (for POST requests)
	if token == "" {
		var req struct {
			Token string `json:"token" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.Warn("Invalid verification request", zap.Error(err))
			middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format", map[string]interface{}{
				"details": err.Error(),
			})
			return
		}
		token = req.Token
	}

	// Verify account
	if appErr := h.authService.VerifyAccount(c, token); appErr != nil {
		h.logger.Error("Account verification failed", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	// Success response using standardized envelope
	responseData := gin.H{
		"status": "verified",
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Account verified successfully", responseData)

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
		middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format", map[string]interface{}{
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
	// Using standardized success envelope
	responseData := gin.H{
		"message": "If an account exists with that email, a verification email has been sent.",
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Verification email sent", responseData)

	h.logger.Info("Resend verification request processed", zap.String("requestID", middlewares.GetRequestID(c)))
}
