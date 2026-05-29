package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"unicode"

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

		// Validate path parameters for traversal / null-byte / control chars
		for _, param := range c.Params {
			if containsDangerousPathChars(param.Value) {
				logger.Warn("Dangerous path parameter detected",
					zap.String("requestID", requestID),
					zap.String("param", param.Key),
					zap.String("path", c.Request.URL.Path))
				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "Bad Request",
					"message":    "Invalid path parameter",
					"request_id": requestID,
				})
				c.Abort()
				return
			}
		}

		// Validate JSON body for POST/PUT/PATCH requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			// Enforce Content-Type: application/json
			contentType := c.GetHeader("Content-Type")
			if c.Request.ContentLength > 0 && !strings.HasPrefix(contentType, "application/json") {
				logger.Warn("Invalid Content-Type for mutating request",
					zap.String("requestID", requestID),
					zap.String("contentType", contentType),
					zap.String("path", c.Request.URL.Path))
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"error":      "Unsupported Media Type",
					"message":    "Content-Type must be application/json",
					"request_id": requestID,
				})
				c.Abort()
				return
			}

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

// containsDangerousPathChars checks path parameter values for traversal,
// null-byte injection, and control character attacks.
func containsDangerousPathChars(value string) bool {
	// Path traversal
	if strings.Contains(value, "..") {
		return true
	}
	// Null byte injection (URL-encoded or literal)
	if strings.Contains(value, "%00") || strings.ContainsRune(value, '\x00') {
		return true
	}
	// Control characters (ASCII 0-31 except common whitespace)
	for _, r := range value {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return true
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
