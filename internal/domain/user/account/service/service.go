package account

import (
	"github.com/gin-gonic/gin"
	accountRepo "github.com/refynehq/refyne-backend/internal/domain/user/account/repository"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type UserAccountService interface {
	CreateDefaultUserSettings(c *gin.Context, userID string) *errors.AppError
	GetUserSettings(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError)
	UpdateUserSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError
}

type UserAccountServiceImpl struct {
	logger      *zap.Logger
	serviceName string

	// Repository Dependencies
	userSettingsRepo accountRepo.UserAccountRepository
}

func NewUserAccountService(userSettingsRepo accountRepo.UserAccountRepository) UserAccountService {
	return &UserAccountServiceImpl{
		logger:           logging.GetServiceLogger("UserAccountService"),
		serviceName:      "UserAccountService",
		userSettingsRepo: userSettingsRepo,
	}
}
