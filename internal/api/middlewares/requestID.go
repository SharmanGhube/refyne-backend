package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID
	RequestIDKey = "request_id"
)

// RequestIDMiddleware generates a unique request ID for each request
// and adds it to both the request context and response headers
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already provided in headers
		requestID := c.GetHeader(RequestIDHeader)

		// If no request ID provided, generate a new one
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to context for use throughout the request lifecycle
		c.Set(RequestIDKey, requestID)

		// Add request ID to response headers for client tracking
		c.Header(RequestIDHeader, requestID)

		// Continue to next handler
		c.Next()
	}
}

// GetRequestID retrieves the request ID from the gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
