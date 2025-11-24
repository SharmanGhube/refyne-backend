package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/shared/validation"
	"go.uber.org/zap"
)

var validator = validation.NewValidator()

// InputValidationMiddleware validates and sanitizes request inputs
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Skip validation for certain paths
		if shouldSkipValidation(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Validate query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if validator.DetectXSS(value) {
					logger.Warn("XSS detected in query parameter",
						zap.String("requestID", requestID),
						zap.String("param", key),
						zap.String("path", c.Request.URL.Path))
					c.JSON(http.StatusBadRequest, gin.H{
						"error":      "Bad Request",
						"message":    "Invalid input detected in query parameters",
						"request_id": requestID,
					})
					c.Abort()
					return
				}

				if validator.DetectSQLInjection(value) {
					logger.Warn("SQL injection detected in query parameter",
						zap.String("requestID", requestID),
						zap.String("param", key),
						zap.String("path", c.Request.URL.Path))
					c.JSON(http.StatusBadRequest, gin.H{
						"error":      "Bad Request",
						"message":    "Invalid input detected in query parameters",
						"request_id": requestID,
					})
					c.Abort()
					return
				}
			}
		}

		// Validate JSON body for POST/PUT/PATCH requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if c.Request.Body != nil {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil && len(bodyBytes) > 0 {
					// Restore body for downstream handlers
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					// Try to parse as JSON
					var jsonBody map[string]interface{}
					if json.Unmarshal(bodyBytes, &jsonBody) == nil {
						// Validate each string value in JSON
						if detectDangerousPatterns(jsonBody) {
							logger.Warn("Dangerous patterns detected in request body",
								zap.String("requestID", requestID),
								zap.String("path", c.Request.URL.Path))
							c.JSON(http.StatusBadRequest, gin.H{
								"error":      "Bad Request",
								"message":    "Invalid input detected in request body",
								"request_id": requestID,
							})
							c.Abort()
							return
						}
					}
				}
			}
		}

		c.Next()
	}
}

// shouldSkipValidation checks if path should skip validation
func shouldSkipValidation(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/api/health",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath || len(path) > len(skipPath) && path[:len(skipPath)+1] == skipPath+"/" {
			return true
		}
	}

	return false
}

// detectDangerousPatterns recursively checks JSON for dangerous patterns
func detectDangerousPatterns(data interface{}) bool {
	switch v := data.(type) {
	case string:
		if validator.DetectXSS(v) || validator.DetectSQLInjection(v) {
			return true
		}
	case map[string]interface{}:
		for _, value := range v {
			if detectDangerousPatterns(value) {
				return true
			}
		}
	case []interface{}:
		for _, item := range v {
			if detectDangerousPatterns(item) {
				return true
			}
		}
	}

	return false
}

// ValidateRequestSize middleware validates request body size
func ValidateRequestSize(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			requestID := GetRequestID(c)
			logger.Warn("Request body too large",
				zap.String("requestID", requestID),
				zap.Int64("size", c.Request.ContentLength),
				zap.Int64("max", maxSize))
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":      "Request Entity Too Large",
				"message":    "Request body exceeds maximum allowed size",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
