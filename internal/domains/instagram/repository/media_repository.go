package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// InstagramMediaRepository defines operations for Instagram media
type InstagramMediaRepository interface {
	// Create a new media record
	CreateMedia(c *gin.Context, input *models.CreateInstagramMediaInput) (*models.InstagramMedia, *errors.AppError)

	// Get media by ID
	GetMediaByID(c *gin.Context, mediaID string) (*models.InstagramMedia, *errors.AppError)

	// Get all media for an account
	GetMediaByAccountID(c *gin.Context, accountID string) ([]*models.InstagramMedia, *errors.AppError)

	// Update media
	UpdateMedia(c *gin.Context, mediaID string, input *models.UpdateInstagramMediaInput) *errors.AppError

	// Delete media
	DeleteMedia(c *gin.Context, mediaID string) *errors.AppError

	// Bulk upsert media (create or update) - accepts context.Context for job queue compatibility
	UpsertMedia(ctx context.Context, accountID string, media []*models.CreateInstagramMediaInput) error

	// Get latest media for an account
	GetLatestMedia(c *gin.Context, accountID string, limit int) ([]*models.InstagramMedia, *errors.AppError)

	// Check if media exists
	MediaExists(c *gin.Context, mediaID string) (bool, *errors.AppError)
}

type instagramMediaRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewInstagramMediaRepository creates a new media repository
func NewInstagramMediaRepository(db *sqlx.DB) InstagramMediaRepository {
	return &instagramMediaRepository{
		db:     db,
		logger: logging.GetRepositoryLogger("InstagramMediaRepository"),
	}
}

// CreateMedia creates a new media record
func (r *instagramMediaRepository) CreateMedia(c *gin.Context, input *models.CreateInstagramMediaInput) (*models.InstagramMedia, *errors.AppError) {
	query := `
		INSERT INTO instagram_media (
			account_id, instagram_media_id, media_type, caption, media_url,
			permalink, thumbnail_url, posted_at, like_count, comment_count, synced_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW()
		) RETURNING *
	`

	var media models.InstagramMedia
	err := r.db.QueryRowxContext(c, query,
		input.AccountID,
		input.InstagramMediaID,
		input.MediaType,
		input.Caption,
		input.MediaURL,
		input.Permalink,
		input.ThumbnailURL,
		input.PostedAt,
		input.LikeCount,
		input.CommentCount,
	).StructScan(&media)

	if err != nil {
		r.logger.Error("Failed to create media", zap.Error(err))
		return nil, errors.NewAppError(
			c,
			"MEDIA_CREATE_FAILED",
			"Failed to create media record",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	r.logger.Debug("Media created", zap.String("media_id", media.ID))
	return &media, nil
}

// GetMediaByID retrieves media by ID
func (r *instagramMediaRepository) GetMediaByID(c *gin.Context, mediaID string) (*models.InstagramMedia, *errors.AppError) {
	query := "SELECT * FROM instagram_media WHERE id = $1"

	var media models.InstagramMedia
	err := r.db.QueryRowxContext(c, query, mediaID).StructScan(&media)

	if err != nil {
		r.logger.Warn("Media not found", zap.String("media_id", mediaID), zap.Error(err))
		return nil, errors.NewAppError(
			c,
			"MEDIA_NOT_FOUND",
			"Media not found",
			errors.ErrorTypeNotFound,
			errors.SeverityMedium,
			"instagram",
		)
	}

	return &media, nil
}

// GetMediaByAccountID retrieves all media for an account
func (r *instagramMediaRepository) GetMediaByAccountID(c *gin.Context, accountID string) ([]*models.InstagramMedia, *errors.AppError) {
	query := "SELECT * FROM instagram_media WHERE account_id = $1 ORDER BY posted_at DESC"

	var media []*models.InstagramMedia
	err := r.db.SelectContext(c, &media, query, accountID)

	if err != nil {
		r.logger.Error("Failed to get media for account", zap.Error(err), zap.String("account_id", accountID))
		return nil, errors.NewAppError(
			c,
			"MEDIA_FETCH_FAILED",
			"Failed to fetch media",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return media, nil
}

// UpdateMedia updates media information
func (r *instagramMediaRepository) UpdateMedia(c *gin.Context, mediaID string, input *models.UpdateInstagramMediaInput) *errors.AppError {
	// Marshal AIAnalysis to JSON if present
	var aiAnalysisJSON *string
	if input.AIAnalysis != nil {
		data, err := json.Marshal(input.AIAnalysis)
		if err != nil {
			r.logger.Error("Failed to marshal AI analysis", zap.Error(err))
			return errors.NewAppError(
				c,
				"AI_ANALYSIS_MARSHAL_FAILED",
				"Failed to process AI analysis",
				errors.ErrorTypeInternal,
				errors.SeverityHigh,
				"instagram",
			)
		}
		jsonStr := string(data)
		aiAnalysisJSON = &jsonStr
	}

	query := `
		UPDATE instagram_media
		SET like_count = $1, comment_count = $2, shares_count = $3, impressions = $4, reach = $5, ai_analysis = $6, synced_at = $7, updated_at = NOW()
		WHERE id = $8
	`

	result, err := r.db.ExecContext(c, query,
		input.LikeCount,
		input.CommentCount,
		input.SharesCount,
		input.Impressions,
		input.Reach,
		aiAnalysisJSON,
		input.SyncedAt,
		mediaID,
	)

	if err != nil {
		r.logger.Error("Failed to update media", zap.Error(err), zap.String("media_id", mediaID))
		return errors.NewAppError(
			c,
			"MEDIA_UPDATE_FAILED",
			"Failed to update media",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return errors.NewAppError(
			c,
			"MEDIA_NOT_FOUND",
			"Media not found",
			errors.ErrorTypeNotFound,
			errors.SeverityMedium,
			"instagram",
		)
	}

	r.logger.Debug("Media updated", zap.String("media_id", mediaID))
	return nil
}

// DeleteMedia deletes a media record
func (r *instagramMediaRepository) DeleteMedia(c *gin.Context, mediaID string) *errors.AppError {
	query := "DELETE FROM instagram_media WHERE id = $1"

	result, err := r.db.ExecContext(c, query, mediaID)
	if err != nil {
		r.logger.Error("Failed to delete media", zap.Error(err), zap.String("media_id", mediaID))
		return errors.NewAppError(
			c,
			"MEDIA_DELETE_FAILED",
			"Failed to delete media",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.NewAppError(
			c,
			"MEDIA_NOT_FOUND",
			"Media not found",
			errors.ErrorTypeNotFound,
			errors.SeverityMedium,
			"instagram",
		)
	}

	return nil
}

// UpsertMedia performs bulk upsert (insert or update)
func (r *instagramMediaRepository) UpsertMedia(ctx context.Context, accountID string, media []*models.CreateInstagramMediaInput) error {
	if len(media) == 0 {
		return nil
	}

	query := `
		INSERT INTO instagram_media (
			account_id, instagram_media_id, media_type, caption, media_url,
			permalink, posted_at, like_count, comment_count, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		ON CONFLICT (instagram_media_id) DO UPDATE SET
			caption = EXCLUDED.caption,
			like_count = EXCLUDED.like_count,
			comment_count = EXCLUDED.comment_count,
			updated_at = NOW()
	`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for _, m := range media {
		_, err := tx.ExecContext(ctx, query,
			accountID,
			m.InstagramMediaID,
			m.MediaType,
			m.Caption,
			m.MediaURL,
			m.Permalink,
			m.PostedAt,
			m.LikeCount,
			m.CommentCount,
		)
		if err != nil {
			tx.Rollback()
			r.logger.Error("Failed to upsert media", zap.Error(err), zap.String("media_id", m.InstagramMediaID))
			return fmt.Errorf("failed to upsert media %s: %w", m.InstagramMediaID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Media upserted", zap.String("account_id", accountID), zap.Int("count", len(media)))
	return nil
}

// GetLatestMedia retrieves the latest media for an account
func (r *instagramMediaRepository) GetLatestMedia(c *gin.Context, accountID string, limit int) ([]*models.InstagramMedia, *errors.AppError) {
	query := fmt.Sprintf(
		"SELECT * FROM instagram_media WHERE account_id = $1 ORDER BY posted_at DESC LIMIT %d",
		limit,
	)

	var media []*models.InstagramMedia
	err := r.db.SelectContext(c, &media, query, accountID)

	if err != nil {
		r.logger.Error("Failed to get latest media", zap.Error(err), zap.String("account_id", accountID))
		return nil, errors.NewAppError(
			c,
			"MEDIA_FETCH_FAILED",
			"Failed to fetch latest media",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return media, nil
}

// MediaExists checks if media exists
func (r *instagramMediaRepository) MediaExists(c *gin.Context, mediaID string) (bool, *errors.AppError) {
	var exists bool
	err := r.db.QueryRowContext(c, "SELECT EXISTS(SELECT 1 FROM instagram_media WHERE id = $1)", mediaID).Scan(&exists)

	if err != nil {
		r.logger.Error("Failed to check media existence", zap.Error(err), zap.String("media_id", mediaID))
		return false, errors.NewAppError(
			c,
			"MEDIA_CHECK_FAILED",
			"Failed to check media",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"instagram",
		)
	}

	return exists, nil
}
