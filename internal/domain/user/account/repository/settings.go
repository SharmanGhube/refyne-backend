package account

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	repoErrors "github.com/refynehq/refyne-backend/internal/domain/user/account/repository/errors"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

func (r *userSettingsRepository) CreateUserSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError {
	r.logger.Info("Creating user settings",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", settings.UserID))

	// Set timestamps if not already set
	if settings.CreatedAt == nil {
		now := time.Now()
		settings.CreatedAt = &now
		settings.UpdatedAt = &now
	}

	// Execute the insert query
	_, err := r.db.NamedExecContext(c.Request.Context(), insertUserSettingsQuery, settings)
	if err != nil {
		r.logger.Error("Failed to create user settings", zap.Error(err))
		return repoErrors.NewUserSettingsDatabaseError(c, "CreateUserSettings", err)
	}

	r.logger.Info("User settings created successfully", zap.String("userID", settings.UserID))
	return nil
}

func (r *userSettingsRepository) GetUserSettingsByUserID(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError) {
	r.logger.Info("Getting user settings by user ID",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", userID))

	var settings userModels.UserSettings
	err := r.db.GetContext(c.Request.Context(), &settings, selectUserSettingsByUserIDQuery, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repoErrors.NewUserSettingsNotFoundError(c, userID)
		}
		r.logger.Error("Failed to get user settings by user ID", zap.Error(err))
		return nil, repoErrors.NewUserSettingsDatabaseError(c, "GetUserSettingsByUserID", err)
	}

	return &settings, nil
}

func (r *userSettingsRepository) UpdateUserSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError {
	r.logger.Info("Updating user settings",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", settings.UserID))

	// Set updated timestamp
	now := time.Now()
	settings.UpdatedAt = &now

	// Execute the update query
	result, err := r.db.NamedExecContext(c.Request.Context(), updateUserSettingsQuery, settings)
	if err != nil {
		r.logger.Error("Failed to update user settings", zap.Error(err))
		return repoErrors.NewUserSettingsDatabaseError(c, "UpdateUserSettings", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return repoErrors.NewUserSettingsDatabaseError(c, "UpdateUserSettings-RowsAffected", err)
	}

	if rowsAffected == 0 {
		return repoErrors.NewUserSettingsNotFoundError(c, settings.UserID)
	}

	r.logger.Info("User settings updated successfully", zap.String("userID", settings.UserID))
	return nil
}

func (r *userSettingsRepository) DeleteUserSettings(c *gin.Context, userID string) *errors.AppError {
	r.logger.Info("Deleting user settings",
		zap.String("requestID", middlewares.GetRequestID(c)),
		zap.String("userID", userID))

	// Execute the delete query
	result, err := r.db.ExecContext(c.Request.Context(), deleteUserSettingsQuery, userID)
	if err != nil {
		r.logger.Error("Failed to delete user settings", zap.Error(err))
		return repoErrors.NewUserSettingsDatabaseError(c, "DeleteUserSettings", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return repoErrors.NewUserSettingsDatabaseError(c, "DeleteUserSettings-RowsAffected", err)
	}

	if rowsAffected == 0 {
		return repoErrors.NewUserSettingsNotFoundError(c, userID)
	}

	r.logger.Info("User settings deleted successfully", zap.String("userID", userID))
	return nil
}
