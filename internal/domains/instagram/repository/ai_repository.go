package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// InstagramAIRepository defines operations for AI recommendations
type InstagramAIRepository interface {
	// Store AI recommendations for media
	StoreRecommendations(ctx context.Context, recommendations *models.AIRecommendations) error

	// Get recommendations for media
	GetRecommendations(ctx context.Context, mediaID string) (*models.AIRecommendations, error)

	// Get recommendations for all media in account
	GetAccountRecommendations(ctx context.Context, accountID string, limit int) ([]*models.AIRecommendations, error)

	// Check if recommendations exist for media
	RecommendationsExist(ctx context.Context, mediaID string) (bool, error)
}

type instagramAIRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewInstagramAIRepository creates a new AI repository
func NewInstagramAIRepository(db *sqlx.DB) InstagramAIRepository {
	return &instagramAIRepository{
		db:     db,
		logger: logging.GetRepositoryLogger("InstagramAIRepository"),
	}
}

// StoreRecommendations stores AI recommendations for media
func (r *instagramAIRepository) StoreRecommendations(ctx context.Context, recommendations *models.AIRecommendations) error {
	// Marshal JSON fields
	analysisJSON, err := json.Marshal(recommendations.Analysis)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	captionsJSON, err := json.Marshal(recommendations.CaptionSuggestions)
	if err != nil {
		return fmt.Errorf("failed to marshal captions: %w", err)
	}

	strategyJSON, err := json.Marshal(recommendations.PostingStrategy)
	if err != nil {
		return fmt.Errorf("failed to marshal strategy: %w", err)
	}

	query := `
		INSERT INTO instagram_ai_recommendations (
			media_id, account_id, analysis, caption_suggestions, posting_strategy,
			confidence_score, generated_at, expires_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7, NOW())
		ON CONFLICT (media_id) DO UPDATE SET
			analysis = EXCLUDED.analysis,
			caption_suggestions = EXCLUDED.caption_suggestions,
			posting_strategy = EXCLUDED.posting_strategy,
			confidence_score = EXCLUDED.confidence_score,
			expires_at = EXCLUDED.expires_at,
			updated_at = NOW()
	`

	_, err = r.db.ExecContext(ctx, query,
		recommendations.MediaID,
		recommendations.AccountID,
		analysisJSON,
		captionsJSON,
		strategyJSON,
		recommendations.ConfidenceScore,
		recommendations.ExpiresAt,
	)

	if err != nil {
		r.logger.Error("Failed to store AI recommendations", zap.Error(err), zap.String("media_id", recommendations.MediaID))
		return fmt.Errorf("failed to store recommendations: %w", err)
	}

	r.logger.Debug("AI recommendations stored", zap.String("media_id", recommendations.MediaID))
	return nil
}

// GetRecommendations retrieves recommendations for media
func (r *instagramAIRepository) GetRecommendations(ctx context.Context, mediaID string) (*models.AIRecommendations, error) {
	query := `
		SELECT id, media_id, account_id, analysis, caption_suggestions, posting_strategy,
			confidence_score, generated_at, expires_at, updated_at
		FROM instagram_ai_recommendations
		WHERE media_id = $1 AND expires_at > NOW()
		LIMIT 1
	`

	var rec models.AIRecommendations
	var analysisJSON, captionsJSON, strategyJSON string

	err := r.db.QueryRowxContext(ctx, query, mediaID).Scan(
		&rec.ID,
		&rec.MediaID,
		&rec.AccountID,
		&analysisJSON,
		&captionsJSON,
		&strategyJSON,
		&rec.ConfidenceScore,
		&rec.GeneratedAt,
		&rec.ExpiresAt,
		&rec.UpdatedAt,
	)

	if err != nil {
		r.logger.Debug("Recommendations not found", zap.String("media_id", mediaID))
		return nil, fmt.Errorf("recommendations not found")
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(analysisJSON), &rec.Analysis); err != nil {
		r.logger.Error("Failed to unmarshal analysis", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal analysis: %w", err)
	}

	if err := json.Unmarshal([]byte(captionsJSON), &rec.CaptionSuggestions); err != nil {
		r.logger.Error("Failed to unmarshal captions", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal captions: %w", err)
	}

	if err := json.Unmarshal([]byte(strategyJSON), &rec.PostingStrategy); err != nil {
		r.logger.Error("Failed to unmarshal strategy", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal strategy: %w", err)
	}

	return &rec, nil
}

// GetAccountRecommendations retrieves recommendations for all media in account
func (r *instagramAIRepository) GetAccountRecommendations(ctx context.Context, accountID string, limit int) ([]*models.AIRecommendations, error) {
	query := `
		SELECT id, media_id, account_id, analysis, caption_suggestions, posting_strategy,
			confidence_score, generated_at, expires_at, updated_at
		FROM instagram_ai_recommendations
		WHERE account_id = $1 AND expires_at > NOW()
		ORDER BY generated_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryxContext(ctx, query, accountID, limit)
	if err != nil {
		r.logger.Error("Failed to query recommendations", zap.Error(err))
		return nil, fmt.Errorf("failed to query recommendations: %w", err)
	}
	defer rows.Close()

	var recommendations []*models.AIRecommendations
	for rows.Next() {
		var rec models.AIRecommendations
		var analysisJSON, captionsJSON, strategyJSON string

		err := rows.Scan(
			&rec.ID,
			&rec.MediaID,
			&rec.AccountID,
			&analysisJSON,
			&captionsJSON,
			&strategyJSON,
			&rec.ConfidenceScore,
			&rec.GeneratedAt,
			&rec.ExpiresAt,
			&rec.UpdatedAt,
		)

		if err != nil {
			r.logger.Error("Failed to scan recommendation row", zap.Error(err))
			continue
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(analysisJSON), &rec.Analysis); err == nil {
			if err := json.Unmarshal([]byte(captionsJSON), &rec.CaptionSuggestions); err == nil {
				if err := json.Unmarshal([]byte(strategyJSON), &rec.PostingStrategy); err == nil {
					recommendations = append(recommendations, &rec)
				}
			}
		}
	}

	return recommendations, nil
}

// RecommendationsExist checks if recommendations exist for media
func (r *instagramAIRepository) RecommendationsExist(ctx context.Context, mediaID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM instagram_ai_recommendations WHERE media_id = $1 AND expires_at > NOW())`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, mediaID).Scan(&exists)
	if err != nil {
		r.logger.Error("Failed to check recommendations existence", zap.Error(err))
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists, nil
}
