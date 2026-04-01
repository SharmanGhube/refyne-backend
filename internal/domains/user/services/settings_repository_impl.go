package user

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// Settings queries
const (
	selectSettingsByUserIDQuery = `
		SELECT id, user_id, language, time_zone, email_notifications, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`

	insertSettingsQuery = `
		INSERT INTO user_settings (id, user_id, language, time_zone, email_notifications, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	upsertSettingsQuery = `
		INSERT INTO user_settings (id, user_id, language, time_zone, email_notifications, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id)
		DO UPDATE SET
			language = EXCLUDED.language,
			time_zone = EXCLUDED.time_zone,
			email_notifications = EXCLUDED.email_notifications,
			updated_at = EXCLUDED.updated_at
	`

	deleteSettingsQuery = `
		DELETE FROM user_settings WHERE user_id = $1
	`
)

// SettingsRepositoryImpl implements SettingsRepository
type SettingsRepositoryImpl struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *sqlx.DB) SettingsRepository {
	return &SettingsRepositoryImpl{
		name:   "SettingsRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("SettingsRepository"),
	}
}

// GetSettings retrieves user settings by user ID
func (r *SettingsRepositoryImpl) GetSettings(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError) {
	r.logger.Debug("Getting user settings",
		zap.String("userID", userID),
		zap.String("requestID", middlewares.GetRequestID(c)))

	var settings userModels.UserSettings
	err := r.db.GetContext(c.Request.Context(), &settings, selectSettingsByUserIDQuery, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("User settings not found",
				zap.String("userID", userID),
				zap.String("requestID", middlewares.GetRequestID(c)))
			return nil, nil // Return nil to indicate no settings exist
		}
		r.logger.Error("Failed to get user settings",
			zap.Error(err),
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return nil, errors.NewAppError(
			c,
			"DATABASE_ERROR",
			"Failed to retrieve user settings",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"settings",
		)
	}

	r.logger.Debug("User settings retrieved",
		zap.String("userID", userID),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return &settings, nil
}

// UpdateSettings updates or creates user settings (upsert)
func (r *SettingsRepositoryImpl) UpdateSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError {
	r.logger.Debug("Updating user settings",
		zap.String("userID", settings.UserID),
		zap.String("requestID", middlewares.GetRequestID(c)))

	now := time.Now()

	// Generate ID if not set
	if settings.ID == "" {
		settings.ID = uuid.New().String()
	}

	// Set timestamps
	if settings.CreatedAt.IsZero() {
		settings.CreatedAt = now
	}
	settings.UpdatedAt = now

	_, err := r.db.ExecContext(c.Request.Context(), upsertSettingsQuery,
		settings.ID,
		settings.UserID,
		settings.Language,
		settings.TimeZone,
		settings.EmailNotifications,
		settings.CreatedAt,
		settings.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to update user settings",
			zap.Error(err),
			zap.String("userID", settings.UserID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return errors.NewAppError(
			c,
			"DATABASE_ERROR",
			"Failed to update user settings",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"settings",
		)
	}

	r.logger.Info("User settings updated",
		zap.String("userID", settings.UserID),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return nil
}

// CreateSettings creates new user settings
func (r *SettingsRepositoryImpl) CreateSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError {
	r.logger.Debug("Creating user settings",
		zap.String("userID", settings.UserID),
		zap.String("requestID", middlewares.GetRequestID(c)))

	now := time.Now()

	// Generate ID if not set
	if settings.ID == "" {
		settings.ID = uuid.New().String()
	}

	settings.CreatedAt = now
	settings.UpdatedAt = now

	_, err := r.db.ExecContext(c.Request.Context(), insertSettingsQuery,
		settings.ID,
		settings.UserID,
		settings.Language,
		settings.TimeZone,
		settings.EmailNotifications,
		settings.CreatedAt,
		settings.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Failed to create user settings",
			zap.Error(err),
			zap.String("userID", settings.UserID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return errors.NewAppError(
			c,
			"DATABASE_ERROR",
			"Failed to create user settings",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"settings",
		)
	}

	r.logger.Info("User settings created",
		zap.String("userID", settings.UserID),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return nil
}

// DeleteSettings deletes user settings
func (r *SettingsRepositoryImpl) DeleteSettings(c *gin.Context, userID string) *errors.AppError {
	r.logger.Debug("Deleting user settings",
		zap.String("userID", userID),
		zap.String("requestID", middlewares.GetRequestID(c)))

	result, err := r.db.ExecContext(c.Request.Context(), deleteSettingsQuery, userID)
	if err != nil {
		r.logger.Error("Failed to delete user settings",
			zap.Error(err),
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return errors.NewAppError(
			c,
			"DATABASE_ERROR",
			"Failed to delete user settings",
			errors.ErrorTypeInternal,
			errors.SeverityMedium,
			"settings",
		)
	}

	rowsAffected, _ := result.RowsAffected()
	r.logger.Info("User settings deleted",
		zap.String("userID", userID),
		zap.Int64("rowsAffected", rowsAffected),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return nil
}
