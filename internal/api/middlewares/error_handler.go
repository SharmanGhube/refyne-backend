package middlewares

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

// ErrorResponse is the standardized error response structure
type ErrorResponse struct {
	Success   bool                   `json:"success"`
	Error     string                 `json:"error"`
	Message   string                 `json:"message"`
	Code      string                 `json:"code,omitempty"`
	RequestID string                 `json:"request_id"`
	Timestamp string                 `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse is the standardized success response structure
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp string      `json:"timestamp"`
}

// ErrorHandlerMiddleware centralizes error handling
func ErrorHandlerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with stack trace
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
				)

				// Return standardized error response
				RespondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred", nil)
			}
		}()

		c.Next()

		// Check if there are any errors after handlers executed
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Check if it's an AppError
			if appErr, ok := err.Err.(*errors.AppError); ok {
				HandleAppError(c, appErr)
			} else {
				// Generic error
				logger.Error("Unhandled error",
					zap.Error(err.Err),
					zap.String("path", c.Request.URL.Path),
				)
				RespondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An error occurred", nil)
			}
		}
	}
}

// HandleAppError handles AppError responses
func HandleAppError(c *gin.Context, appErr *errors.AppError) {
	// Don't expose internal error details in production
	message := appErr.Message
	details := make(map[string]interface{})

	// Only include safe details
	if appErr.Type == errors.ErrorTypeValidation {
		details = appErr.Details
	}

	RespondWithError(c, appErr.HTTPStatus, appErr.Code, message, details)
}

// RespondWithError sends a standardized error response
func RespondWithError(c *gin.Context, statusCode int, code, message string, details map[string]interface{}) {
	// Prevent double responses
	if c.Writer.Written() {
		return
	}

	requestID := GetRequestID(c)

	response := ErrorResponse{
		Success:   false,
		Error:     code,
		Message:   message,
		Code:      code,
		RequestID: requestID,
		Timestamp: GetCurrentTimestamp(),
		Details:   details,
	}

	c.JSON(statusCode, response)
	c.Abort()
}

// RespondWithSuccess sends a standardized success response
func RespondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	requestID := GetRequestID(c)

	response := SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID,
		Timestamp: GetCurrentTimestamp(),
	}

	c.JSON(statusCode, response)
}

// Validation error helpers
func RespondWithValidationError(c *gin.Context, field, message string) {
	details := map[string]interface{}{
		"field": field,
	}
	RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", message, details)
}

func RespondWithUnauthorized(c *gin.Context, message string) {
	RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func RespondWithForbidden(c *gin.Context, message string) {
	RespondWithError(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func RespondWithNotFound(c *gin.Context, resource string) {
	message := resource + " not found"
	RespondWithError(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func RespondWithConflict(c *gin.Context, message string) {
	RespondWithError(c, http.StatusConflict, "CONFLICT", message, nil)
}

func RespondWithInternalError(c *gin.Context) {
	RespondWithError(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred", nil)
}
