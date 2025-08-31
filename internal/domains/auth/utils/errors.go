package auth

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

const (
	CodeJWTSecretNotSet = "JWT_SECRET_NOT_SET"

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

	// UUID Generation Errors
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

func NewJWTGenerationError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTGenerationFailed,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
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

func NewUUIDGenerationFailedError(c *gin.Context, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeUUIDGenerationFailed,
		"Failed to generate UUID: ",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"auth",
	).WithContext("error", err.Error())
}

func NewJWTTokenInvalidError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTTokenInvalid,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityMedium,
		"auth",
	)
}

func NewJWTTokenExpiredError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTTokenExpired,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityLow,
		"auth",
	)
}

func NewJWTTokenRevokedError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTTokenRevoked,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityMedium,
		"auth",
	)
}

func NewJWTSigningMethodInvalidError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTSigningMethodInvalid,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityMedium,
		"auth",
	)
}

func NewJWTValidationFailedError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeJWTValidationFailed,
		message,
		errors.ErrorTypeUnauthorized,
		errors.SeverityMedium,
		"auth",
	)
}
