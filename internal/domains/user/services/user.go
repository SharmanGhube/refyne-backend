package user

import (
	"github.com/gin-gonic/gin"
	userRepo "github.com/refynehq/refyne-backend/internal/domains/user/core/repository"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// UserService defines business logic for user operations
type UserService interface {
	// Profile operations
	GetUserProfile(c *gin.Context, userID string) (*userModels.User, *errors.AppError)
	UpdateUserProfile(c *gin.Context, userID string, update *ProfileUpdateRequest) (*userModels.User, *errors.AppError)

	// Settings operations
	GetUserSettings(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError)
	UpdateUserSettings(c *gin.Context, userID string, settings *userModels.UserSettings) (*userModels.UserSettings, *errors.AppError)

	// Onboarding
	CompleteOnboarding(c *gin.Context, userID string) *errors.AppError

	// Account deletion (soft delete)
	DeleteUserAccount(c *gin.Context, userID string) *errors.AppError
}

// UserServiceImpl implements UserService
type UserServiceImpl struct {
	name   string
	logger *zap.Logger

	// Repository dependencies
	userRepo     userRepo.CoreUserRepository
	settingsRepo SettingsRepository
}

// NewUserService creates a new user service
func NewUserService(
	userRepo userRepo.CoreUserRepository,
	settingsRepo SettingsRepository,
) UserService {
	return &UserServiceImpl{
		name:         "UserService",
		logger:       logging.GetServiceLogger("UserService"),
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
	}
}

// ProfileUpdateRequest represents a profile update
type ProfileUpdateRequest struct {
	FirstName string `json:"first_name" binding:"omitempty,max=100"`
	LastName  string `json:"last_name" binding:"omitempty,max=100"`
	Username  string `json:"username" binding:"omitempty,alphanum,min=3,max=50"`
}

// GetUserProfile retrieves the user's profile
func (s *UserServiceImpl) GetUserProfile(c *gin.Context, userID string) (*userModels.User, *errors.AppError) {
	s.logger.Debug("Fetching user profile", zap.String("userID", userID))

	user, appErr := s.userRepo.GetUserByID(c, userID)
	if appErr != nil {
		s.logger.Error("Failed to fetch user profile", zap.Error(appErr))
		return nil, appErr
	}

	if user == nil {
		s.logger.Warn("User not found", zap.String("userID", userID))
		return nil, errors.NewAppError(
			c,
			"USER_NOT_FOUND",
			"User not found",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"user",
		)
	}

	return user, nil
}

// UpdateUserProfile updates the user's profile information
func (s *UserServiceImpl) UpdateUserProfile(c *gin.Context, userID string, update *ProfileUpdateRequest) (*userModels.User, *errors.AppError) {
	s.logger.Debug("Updating user profile", zap.String("userID", userID))

	// Get current user
	user, appErr := s.userRepo.GetUserByID(c, userID)
	if appErr != nil {
		return nil, appErr
	}

	if user == nil {
		return nil, errors.NewAppError(
			c,
			"USER_NOT_FOUND",
			"User not found",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"user",
		)
	}

	// Update fields if provided
	if update.FirstName != "" {
		user.FirstName = update.FirstName
	}
	if update.LastName != "" {
		user.LastName = update.LastName
	}
	if update.Username != "" {
		// Check if username is already taken by another user
		exists, appErr := s.userRepo.UserExistsByUsername(c, update.Username)
		if appErr != nil {
			return nil, appErr
		}
		if exists && user.Username != update.Username {
			return nil, errors.NewAppError(
				c,
				"USERNAME_TAKEN",
				"Username is already in use",
				errors.ErrorTypeValidation,
				errors.SeverityLow,
				"user",
			)
		}
		user.Username = update.Username
	}

	// Persist the changes
	updatedUser, appErr := s.userRepo.UpdateUser(c, user)
	if appErr != nil {
		s.logger.Error("Failed to update user profile", zap.Error(appErr))
		return nil, appErr
	}

	s.logger.Info("User profile updated", zap.String("userID", userID))
	return updatedUser, nil
}

// GetUserSettings retrieves user settings
func (s *UserServiceImpl) GetUserSettings(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError) {
	s.logger.Debug("Fetching user settings", zap.String("userID", userID))

	settings, appErr := s.settingsRepo.GetSettings(c, userID)
	if appErr != nil {
		s.logger.Error("Failed to fetch user settings", zap.Error(appErr))
		return nil, appErr
	}

	// If settings don't exist, create defaults
	if settings == nil {
		settings = &userModels.UserSettings{
			UserID:             userID,
			EmailNotifications: true,
			Language:           "en",
			TimeZone:           "UTC",
		}
	}

	return settings, nil
}

// UpdateUserSettings updates user settings
func (s *UserServiceImpl) UpdateUserSettings(c *gin.Context, userID string, settings *userModels.UserSettings) (*userModels.UserSettings, *errors.AppError) {
	s.logger.Debug("Updating user settings", zap.String("userID", userID))

	settings.UserID = userID

	appErr := s.settingsRepo.UpdateSettings(c, settings)
	if appErr != nil {
		s.logger.Error("Failed to update user settings", zap.Error(appErr))
		return nil, appErr
	}

	s.logger.Info("User settings updated", zap.String("userID", userID))
	return settings, nil
}

// CompleteOnboarding marks onboarding as completed for user
func (s *UserServiceImpl) CompleteOnboarding(c *gin.Context, userID string) *errors.AppError {
	s.logger.Debug("Completing onboarding", zap.String("userID", userID))

	// Get user
	user, appErr := s.userRepo.GetUserByID(c, userID)
	if appErr != nil {
		return appErr
	}

	if user == nil {
		return errors.NewAppError(
			c,
			"USER_NOT_FOUND",
			"User not found",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"user",
		)
	}

	// Persist the onboarding status
	appErr = s.userRepo.UpdateOnboardingStatus(c, userID, true)
	if appErr != nil {
		s.logger.Error("Failed to update onboarding status", zap.Error(appErr))
		return appErr
	}

	s.logger.Info("Onboarding completed", zap.String("userID", userID))
	return nil
}

// DeleteUserAccount performs a soft delete of the user account
func (s *UserServiceImpl) DeleteUserAccount(c *gin.Context, userID string) *errors.AppError {
	s.logger.Debug("Deleting user account", zap.String("userID", userID))

	// Get user
	user, appErr := s.userRepo.GetUserByID(c, userID)
	if appErr != nil {
		return appErr
	}

	if user == nil {
		return errors.NewAppError(
			c,
			"USER_NOT_FOUND",
			"User not found",
			errors.ErrorTypeNotFound,
			errors.SeverityLow,
			"user",
		)
	}

	// Soft delete the user account
	appErr = s.userRepo.SoftDeleteUser(c, userID)
	if appErr != nil {
		s.logger.Error("Failed to delete user account", zap.Error(appErr))
		return appErr
	}

	s.logger.Info("User account deleted", zap.String("userID", userID))
	return nil
}
