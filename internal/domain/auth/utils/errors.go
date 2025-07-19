package auth

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

const (
	// JWT Errors
	CodeJWTSecretNotSet = "JWT_SECRET_KEY_NOT_SET"

	// Generation Errors
	CodeJWTGenerationFailed = "JWT_GENERATION_FAILED"
	CodeJWTClaimsInvalid    = "JWT_CLAIMS_INVALID"

	// Validation Errors
	// JWT Validation errors
	CodeJWTTokenInvalid         = "AUTH_TOKEN_JWT_TOKEN_INVALID"
	CodeJWTTokenExpired         = "AUTH_TOKEN_JWT_TOKEN_EXPIRED"
	CodeJWTTokenRevoked         = "AUTH_TOKEN_JWT_TOKEN_REVOKED"
	CodeJWTValidationFailed     = "AUTH_TOKEN_JWT_VALIDATION_FAILED"
	CodeJWTSigningMethodInvalid = "AUTH_TOKEN_JWT_SIGNING_METHOD_INVALID"

	// UUID Errors
	CodeUUIDGenerationFailed = "UUID_GENERATION_FAILED"
)

func NewJWTSecretNotSetError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTSecretNotSet,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	)
}

func NewJWTGenerationFailedError(c *gin.Context, error error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTGenerationFailed,
		"Failed to generate JWT token",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	).WithContext(
		"error", error.Error(),
	)
}

func NewJWTClaimsInvalidError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTClaimsInvalid,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	)
}

func NewUUIDGenerationFailedError(c *gin.Context, error error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUUIDGenerationFailed,
		"Failed to generate UUID",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	).WithContext(
		"error", error.Error(),
	)
}

func NewJWTTokenInvalidError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTTokenInvalid,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityHigh,
		"auth",
	)
}

func NewJWTTokenExpiredError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTTokenExpired,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityHigh,
		"auth",
	)
}

func NewJWTTokenRevokedError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTTokenRevoked,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityHigh,
		"auth",
	)
}

func NewJWTValidationFailedError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTValidationFailed,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityHigh,
		"auth",
	)
}

func NewJWTSigningMethodInvalidError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTSigningMethodInvalid,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	)
}
