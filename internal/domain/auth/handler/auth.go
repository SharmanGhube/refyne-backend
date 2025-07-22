package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"go.uber.org/zap"
)

// Request struct for user registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// Handler for user registration
func (h *authHandler) Register(c *gin.Context) {
	var req RegisterRequest

	// Bind and validate JSON request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid registration request",
			zap.Error(err),
			zap.String("request_id", middlewares.GetRequestID(c)))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Processing user registration",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
		zap.String("request_id", middlewares.GetRequestID(c)))

	// Call the auth service to register the user
	if appErr := h.authService.RegisterUser(c, req.Username, req.Password, req.Email); appErr != nil {
		h.logger.Warn("User registration failed",
			zap.String("error", appErr.Message),
			zap.String("code", appErr.Code),
			zap.String("username", req.Username),
			zap.String("email", req.Email))

		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	h.logger.Info("User registration successful",
		zap.String("username", req.Username),
		zap.String("email", req.Email))

	// Success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
	})
}

// Request struct for login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Handler for user login
func (h *authHandler) Login(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "User logged in successfully",
	})
}
