package account

import (
	"github.com/gin-gonic/gin"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

func (s *UserAccountServiceImpl) CreateDefaultUserSettings(c *gin.Context, userID string) *errors.AppError {
	s.logger.Info("Creating default user settings", zap.String("userID", userID))

	// Create default user settings
	defaultSettings := userModels.NewDefaultUserSettings(userID)

	// Save to database
	if appErr := s.userSettingsRepo.CreateUserSettings(c, defaultSettings); appErr != nil {
		s.logger.Error("Failed to create default user settings",
			zap.String("userID", userID),
			zap.Error(appErr))
		return appErr
	}

	s.logger.Info("Default user settings created successfully", zap.String("userID", userID))
	return nil
}

func (s *UserAccountServiceImpl) GetUserSettings(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError) {
	s.logger.Info("Getting user settings", zap.String("userID", userID))

	settings, appErr := s.userSettingsRepo.GetUserSettingsByUserID(c, userID)
	if appErr != nil {
		s.logger.Error("Failed to get user settings",
			zap.String("userID", userID),
			zap.Error(appErr))
		return nil, appErr
	}

	return settings, nil
}

func (s *UserAccountServiceImpl) UpdateUserSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError {
	s.logger.Info("Updating user settings", zap.String("userID", settings.UserID))

	if appErr := s.userSettingsRepo.UpdateUserSettings(c, settings); appErr != nil {
		s.logger.Error("Failed to update user settings",
			zap.String("userID", settings.UserID),
			zap.Error(appErr))
		return appErr
	}

	s.logger.Info("User settings updated successfully", zap.String("userID", settings.UserID))
	return nil
}
