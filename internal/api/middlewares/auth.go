package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	authUtils "github.com/refynehq/refyne-backend/internal/domains/auth/utils"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

const (
	// AuthorizationHeader is the header name for authorization
	AuthorizationHeader = "Authorization"
	// UserIDKey is the context key for authenticated user ID
	UserIDKey = "user_id"
	// UserEmailKey is the context key for authenticated user email
	UserEmailKey = "user_email"
	// UsernameKey is the context key for authenticated username
	UsernameKey = "username"
	// TokenKey is the context key for the token itself
	TokenKey = "token"
)

var logger = logging.GetLogger()

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Extract token from Authorization header
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			logger.Warn("Missing authorization header",
				zap.String("requestID", requestID),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Unauthorized",
				"message":    "Missing authorization token",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format",
				zap.String("requestID", requestID),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Unauthorized",
				"message":    "Invalid authorization format. Expected: Bearer <token>",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		token := parts[1]

		// Check if token is blacklisted
		tokenManager := authUtils.GetTokenBlacklistManager()
		if tokenManager.IsBlacklisted(token) {
			logger.Warn("Attempted use of blacklisted token",
				zap.String("requestID", requestID),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Unauthorized",
				"message":    "Token has been revoked",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		// Validate token and extract claims
		claims, err := authUtils.ValidateAndExtractToken(token)
		if err != nil {
			logger.Warn("Token validation failed",
				zap.String("requestID", requestID),
				zap.String("path", c.Request.URL.Path),
				zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":      "Unauthorized",
				"message":    "Invalid or expired token",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		// Set user information in context for downstream handlers
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UsernameKey, claims.Username)
		c.Set(TokenKey, token)

		logger.Debug("User authenticated successfully",
			zap.String("requestID", requestID),
			zap.String("userID", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("path", c.Request.URL.Path))

		// Continue to next handler
		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens if present but doesn't require them
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)

		// If no auth header, continue without authentication
		if authHeader == "" {
			c.Next()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		token := parts[1]

		// Check if token is blacklisted
		tokenManager := authUtils.GetTokenBlacklistManager()
		if tokenManager.IsBlacklisted(token) {
			c.Next()
			return
		}

		// Validate token and extract claims
		claims, err := authUtils.ValidateAndExtractToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Set user information in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UsernameKey, claims.Username)
		c.Set(TokenKey, token)

		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	id, ok := userID.(string)
	return id, ok
}

// GetUserEmail retrieves the authenticated user email from context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(UserEmailKey)
	if !exists {
		return "", false
	}
	userEmail, ok := email.(string)
	return userEmail, ok
}

// GetUsername retrieves the authenticated username from context
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}

// GetToken retrieves the JWT token from context
func GetToken(c *gin.Context) (string, bool) {
	token, exists := c.Get(TokenKey)
	if !exists {
		return "", false
	}
	tokenStr, ok := token.(string)
	return tokenStr, ok
}

// RequireAuth is a helper to check if user is authenticated
// Returns user ID or aborts with 401
func RequireAuth(c *gin.Context) (string, bool) {
	userID, exists := GetUserID(c)
	if !exists || userID == "" {
		requestID := GetRequestID(c)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "Unauthorized",
			"message":    "Authentication required",
			"request_id": requestID,
		})
		c.Abort()
		return "", false
	}
	return userID, true
}
