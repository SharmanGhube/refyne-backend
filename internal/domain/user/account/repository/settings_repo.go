package account

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// UserSettingsRepository handles CRUD operations for user settings
type UserSettingsRepository interface {
	CreateUserSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError
	GetUserSettingsByUserID(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError)
	UpdateUserSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError
	DeleteUserSettings(c *gin.Context, userID string) *errors.AppError
}

type userSettingsRepository struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

func NewUserSettingsRepository(db *sqlx.DB) UserSettingsRepository {
	return &userSettingsRepository{
		name:   "UserSettingsRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("UserSettingsRepository"),
	}
}
