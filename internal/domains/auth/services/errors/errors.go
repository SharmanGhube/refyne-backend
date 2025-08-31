package auth

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

const (
	CodeUserAlreadyExists  = "USER_ALREADY_EXISTS"
	CodeUserNotFound       = "USER_NOT_FOUND"
	CodeInvalidPassword    = "INVALID_PASSWORD"
	CodeInvalidEmailFormat = "INVALID_EMAIL_FORMAT"

	CodePasswordHashingFailed = "PASSWORD_HASHING_FAILED"

	CodeUserCreationFailed = "NEW_USER_CREATION_FAILED"

	// Status errors
	CodeUserNotActive   = "USER_NOT_ACTIVE"
	CodeUserNotVerified = "USER_NOT_VERIFIED"

	// OTP errors
	CodeInvalidOTP      = "INVALID_OTP"
	CodeOTPNotFound     = "OTP_NOT_FOUND"
	CodeOTPExpired      = "OTP_EXPIRED"
	CodeInternalServer  = "INTERNAL_SERVER_ERROR"

	// Token errors
	CodeInvalidToken           = "INVALID_TOKEN"
	CodeTokenExpired           = "TOKEN_EXPIRED"
	CodeTokenNotFound          = "TOKEN_NOT_FOUND"
	CodeAccountAlreadyVerified = "ACCOUNT_ALREADY_VERIFIED"
)

func NewUserAlreadyExistsError(c *gin.Context, field, value string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserAlreadyExists,
		"User with this "+field+" already exists",
		errors.ErrorTypeConflict,
		errors.SeverityMedium,
		"auth",
	).WithContext(field, value)
}

func NewInvalidPasswordError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidPassword,
		message,
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	)
}

func NewInvalidEmailFormatError(c *gin.Context, email string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidEmailFormat,
		"Invalid email format",
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	).WithContext("email", email)
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

func NewUserCreationFailedError(c *gin.Context, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserCreationFailed,
		"Failed to create new user",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	).WithContext("error", err.Error())
}

func NewUserNotFoundError(c *gin.Context, email string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserNotFound,
		"User with this email does not exist",
		errors.ErrorTypeNotFound,
		errors.SeverityLow,
		"auth",
	).WithContext("email", email)
}

func NewUserNotActiveError(c *gin.Context, email string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserNotActive,
		"User account is not active",
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	).WithContext("email", email)
}

func NewUserNotVerifiedError(c *gin.Context, email string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUserNotVerified,
		"User account is not verified",
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	).WithContext("email", email)
}

func NewInvalidOTPError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidOTP,
		message,
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	)
}

func NewInternalServerError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInternalServer,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	)
}

func NewOTPNotFoundError(c *gin.Context, email string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeOTPNotFound,
		"No OTP found for this email",
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	).WithContext("email", email)
}

func NewOTPExpiredError(c *gin.Context, email string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeOTPExpired,
		"OTP has expired",
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	).WithContext("email", email)
}

func NewInvalidTokenError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidToken,
		message,
		errors.ErrorTypeValidation,
		errors.SeverityMedium,
		"auth",
	)
}

func NewTokenExpiredError(c *gin.Context) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeTokenExpired,
		"Token has expired",
		errors.ErrorTypeValidation,
		errors.SeverityMedium,
		"auth",
	)
}

func NewTokenNotFoundError(c *gin.Context) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeTokenNotFound,
		"Token not found in request",
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	)
}

func NewAccountAlreadyVerifiedError(c *gin.Context, email string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeAccountAlreadyVerified,
		"Account is already verified",
		errors.ErrorTypeValidation,
		errors.SeverityLow,
		"auth",
	).WithContext("email", email)
}
