package account

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// UserAccountRepository handles CRUD operations for user accounts
type UserAccountRepository interface {
	// CRUD operations for user settings
	CreateUserSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError
	GetUserSettingsByUserID(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError)
	UpdateUserSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError
	DeleteUserSettings(c *gin.Context, userID string) *errors.AppError

	// Additional methods can be added as needed
}

type userSettingsRepository struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

func NewUserSettingsRepository(db *sqlx.DB) UserAccountRepository {
	return &userSettingsRepository{
		name:   "UserSettingsRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("UserSettingsRepository"),
	}
}
