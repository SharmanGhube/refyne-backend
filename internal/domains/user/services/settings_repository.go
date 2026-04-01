package user

import (
	"github.com/gin-gonic/gin"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// SettingsRepository defines operations for user settings
type SettingsRepository interface {
	GetSettings(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError)
	UpdateSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError
	CreateSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError
	DeleteSettings(c *gin.Context, userID string) *errors.AppError
}
