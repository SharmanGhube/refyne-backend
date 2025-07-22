package auth

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

const (
	// User Registration Errors
	CodeUserAlreadyExists     = "USER_ALREADY_EXISTS"
	CodeInvalidPassword       = "INVALID_PASSWORD"
	CodeInvalidEmail          = "INVALID_EMAIL"
	CodeUserCreationFailed    = "USER_CREATION_FAILED"
	CodePasswordHashingFailed = "PASSWORD_HASHING_FAILED"
)

func NewUserAlreadyExistsError(c *gin.Context, field, value string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserAlreadyExists,
		"User with this "+field+" already exists",
		errors.ErrorTypeConflict,
		errors.SeverityMedium,
		"auth",
	).WithContext("field", field).WithContext("value", value)
}

func NewInvalidPasswordError(c *gin.Context, reason string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidPassword,
		reason,
		errors.ErrorTypeValidation,
		errors.SeverityMedium,
		"auth",
	)
}

func NewInvalidEmailError(c *gin.Context, email string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidEmail,
		"Invalid email format",
		errors.ErrorTypeValidation,
		errors.SeverityMedium,
		"auth",
	).WithContext("email", email)
}

func NewUserCreationFailedError(c *gin.Context, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserCreationFailed,
		"Failed to create user",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	).WithContext("error", err.Error())
}

func NewPasswordHashingFailedError(c *gin.Context, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodePasswordHashingFailed,
		"Failed to hash password",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	).WithContext("error", err.Error())
}
