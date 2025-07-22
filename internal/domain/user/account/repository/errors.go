package account

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

const (
	// User settings error codes
	CodeUserSettingsNotFound = "USER_SETTINGS_NOT_FOUND"
	CodeUserSettingsDatabase = "USER_SETTINGS_DATABASE_ERROR"
)

func NewUserSettingsNotFoundError(c *gin.Context, userID string) *errors.AppError {
	appErr := errors.NewAppError(
		c,
		CodeUserSettingsNotFound,
		"User settings not found",
		errors.ErrorTypeNotFound,
		errors.SeverityMedium,
		"user-account",
	).WithContext("user_id", userID)
	appErr.HTTPStatus = 404
	return appErr
}

func NewUserSettingsDatabaseError(c *gin.Context, operation string, err error) *errors.AppError {
	appErr := errors.NewAppError(
		c,
		CodeUserSettingsDatabase,
		"Database operation failed",
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"user-account",
	).WithContext("operation", operation).WithContext("error", err.Error())
	appErr.HTTPStatus = 500
	return appErr
}
