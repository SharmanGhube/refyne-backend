package errors

import (
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeInternal     ErrorType = "internal"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	ErrorTypeRateLimit    ErrorType = "rate_limit"
	ErrorTypeConflict     ErrorType = "conflict"
)

type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

type AppError struct {
	Code       string    `json:"code"`
	Message    string    `json:"message"`
	Type       ErrorType `json:"type"`
	Timestamp  time.Time `json:"timestamp"`
	RequestID  string    `json:"request_id"`
	HTTPStatus int       `json:"http_status"`

	// Fields not exposed to the user
	Severity  Severity               `json:"-"`
	Service   string                 `json:"-"`
	Operation string                 `json:"-"`
	Details   map[string]interface{} `json:"-"`

	OriginalError error  `json:"-"`
	StackTrace    string `json:"-"`
}

func (e *AppError) Error() string {
	if e.OriginalError != nil {
		return e.OriginalError.Error()
	}
	return e.Message
}

// WithContext adds additional context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

func NewAppError(c *gin.Context, code string, message string, errType ErrorType, severity Severity, service string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Type:       errType,
		Timestamp:  time.Now(),
		RequestID:  c.GetHeader("request-id"),
		HTTPStatus: 500, // Default to internal server error
		Severity:   severity,
		Service:    service,
		Details:    make(map[string]interface{}),
	}
}

func (e *AppError) WithOperation(operation string) *AppError {
	e.Operation = operation
	return e
}

// ClientResponse formats the error for client response
func (e *AppError) ClientResponse() map[string]interface{} {
	allowedTypes := map[ErrorType]bool{
		ErrorTypeRateLimit:    true,
		ErrorTypeNotFound:     true,
		ErrorTypeUnauthorized: true,
		ErrorTypeConflict:     true,
		ErrorTypeValidation:   true,
	}

	// If error type is not allowed, return a generic error
	if !allowedTypes[e.Type] {
		return map[string]interface{}{
			"error": map[string]interface{}{
				"code":      "INTERNAL_ERROR",
				"message":   "An unexpected error occurred. Please try again later.",
				"type":      ErrorTypeInternal,
				"timestamp": e.Timestamp.Format(time.RFC3339),
			},
			"request_id": e.RequestID,
		}
	}

	response := map[string]any{
		"error": map[string]any{
			"code":      e.Code,
			"message":   e.Message,
			"type":      e.Type,
			"timestamp": e.Timestamp.Format(time.RFC3339),
		},
		"request_id": e.RequestID,
	}

	return response
}
