package middlewares

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// GetCurrentTimestamp returns current time in RFC3339 format
func GetCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

// GetRequestID safely retrieves request ID from context
func GetRequestIDSafe(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return c.GetHeader("X-Request-ID")
}

// LogError logs error with context
func LogError(c *gin.Context, err error, message string) {
	fmt.Printf("[ERROR] %s: %v | Path: %s | Method: %s | RequestID: %s\n",
		message,
		err,
		c.Request.URL.Path,
		c.Request.Method,
		GetRequestIDSafe(c),
	)
}
