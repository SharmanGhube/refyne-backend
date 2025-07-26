package user

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

const (
	// Validation error codes
	CodeUserNotFound         = "USER_NOT_FOUND"
	CodeDuplicateUser        = "DUPLICATE_USER"
	CodeDatabaseError        = "DATABASE_ERROR"
	CodeUserValidationFailed = "USER_VALIDATION_FAILED"
)

func NewUserNotFoundError(c *gin.Context, identifier string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserNotFound,
		"User not found",
		errors.ErrorTypeNotFound,
		errors.SeverityMedium,
		"user",
	).WithContext("identifier", identifier)
}

func NewDuplicateUserError(c *gin.Context, field, value string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeDuplicateUser,
		"User with this "+field+" already exists",
		errors.ErrorTypeConflict,
		errors.SeverityMedium,
		"user",
	).WithContext("field", field).WithContext("value", value)
}

func NewDatabaseError(c *gin.Context, operation string, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeDatabaseError,
		"Database operation failed",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"user",
	).WithContext("operation", operation).WithContext("error", err.Error())
}

func NewUserValidationFailedError(c *gin.Context, field, reason string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserValidationFailed,
		"User validation failed: "+reason,
		errors.ErrorTypeValidation,
		errors.SeverityMedium,
		"user",
	).WithContext("field", field)
}
