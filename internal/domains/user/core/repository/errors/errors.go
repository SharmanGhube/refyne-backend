package user

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

const (
	// Validation error codes
	CodeDatabaseError = "DATABASE_ERROR"
	CodeUserNotFound  = "USER_NOT_FOUND"
)

func NewDatabaseError(c *gin.Context, operation string, err error) *errors.AppError {
	return errors.NewAppError(c, CodeDatabaseError, "A database error occurred", errors.ErrorTypeInternal, errors.SeverityHigh, "user").WithContext("operation", operation).WithContext("error", err.Error())
}

func NewUserNotFoundError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(c, CodeUserNotFound, message, errors.ErrorTypeNotFound, errors.SeverityLow, "user")
}
