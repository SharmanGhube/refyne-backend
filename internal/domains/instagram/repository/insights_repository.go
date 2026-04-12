package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// InstagramInsightsRepository defines operations for Instagram insights
type InstagramInsightsRepository interface {
	// Store insights for a media
	StoreMediaInsights(ctx context.Context, insights *models.MediaInsights) error

	// Get insights for a media
	GetMediaInsights(ctx context.Context, mediaID string) (*models.MediaInsights, error)

	// Get latest insights for an account
	GetLatestInsights(ctx context.Context, accountID string, limit int) ([]*models.MediaInsights, error)

	// Store account insights
	StoreAccountInsights(ctx context.Context, insights *models.AccountInsights) error

	// Get account insights by date
	GetAccountInsightsByDate(ctx context.Context, accountID string, date time.Time) (*models.AccountInsights, error)

	// Get account insights trend (last N days)
	GetAccountInsightsTrend(ctx context.Context, accountID string, days int) ([]*models.AccountInsights, error)
}

type instagramInsightsRepository struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewInstagramInsightsRepository creates a new insights repository
func NewInstagramInsightsRepository(db *sqlx.DB) InstagramInsightsRepository {
	return &instagramInsightsRepository{
		db:     db,
		logger: logging.GetRepositoryLogger("InstagramInsightsRepository"),
	}
}

// StoreMediaInsights stores insights for a media
func (r *instagramInsightsRepository) StoreMediaInsights(ctx context.Context, insights *models.MediaInsights) error {
	query := `
		INSERT INTO instagram_insights (
			media_id, account_id, impressions, reach, profile_visits, shares, saves, clicks,
			engagement_rate, metric_date, collected_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		ON CONFLICT (media_id) DO UPDATE SET
			impressions = EXCLUDED.impressions,
			reach = EXCLUDED.reach,
			profile_visits = EXCLUDED.profile_visits,
			shares = EXCLUDED.shares,
			saves = EXCLUDED.saves,
			clicks = EXCLUDED.clicks,
			engagement_rate = EXCLUDED.engagement_rate,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		insights.MediaID,
		insights.AccountID,
		insights.Impressions,
		insights.Reach,
		insights.ProfileVisits,
		insights.Shares,
		insights.Saves,
		insights.Clicks,
		insights.EngagementRate,
		insights.MetricDate,
	)

	if err != nil {
		r.logger.Error("Failed to store media insights", zap.Error(err), zap.String("media_id", insights.MediaID))
		return fmt.Errorf("failed to store media insights: %w", err)
	}

	r.logger.Debug("Media insights stored", zap.String("media_id", insights.MediaID))
	return nil
}

// GetMediaInsights retrieves insights for a media
func (r *instagramInsightsRepository) GetMediaInsights(ctx context.Context, mediaID string) (*models.MediaInsights, error) {
	query := `
		SELECT id, media_id, account_id, impressions, reach, profile_visits, shares, saves, clicks,
			engagement_rate, metric_date, collected_at, updated_at
		FROM instagram_insights
		WHERE media_id = $1
		ORDER BY metric_date DESC
		LIMIT 1
	`

	var insights models.MediaInsights
	err := r.db.QueryRowxContext(ctx, query, mediaID).StructScan(&insights)
	if err != nil {
		r.logger.Warn("Media insights not found", zap.String("media_id", mediaID), zap.Error(err))
		return nil, fmt.Errorf("media insights not found")
	}

	return &insights, nil
}

// GetLatestInsights retrieves latest insights for an account
func (r *instagramInsightsRepository) GetLatestInsights(ctx context.Context, accountID string, limit int) ([]*models.MediaInsights, error) {
	query := `
		SELECT id, media_id, account_id, impressions, reach, profile_visits, shares, saves, clicks,
			engagement_rate, metric_date, collected_at, updated_at
		FROM instagram_media_insights
		WHERE account_id = $1
		ORDER BY metric_date DESC
		LIMIT $2
	`

	var insights []*models.MediaInsights
	err := r.db.SelectContext(ctx, &insights, query, accountID, limit)
	if err != nil {
		r.logger.Error("Failed to get latest insights", zap.Error(err), zap.String("account_id", accountID))
		return nil, fmt.Errorf("failed to get insights: %w", err)
	}

	return insights, nil
}

// StoreAccountInsights stores account-level insights
func (r *instagramInsightsRepository) StoreAccountInsights(ctx context.Context, insights *models.AccountInsights) error {
	query := `
		INSERT INTO instagram_insights (
			account_id, impressions, reach, profile_views, follower_count,
			engagement_rate, growth_rate, metric_date
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (account_id, metric_date) DO UPDATE SET
			impressions = EXCLUDED.impressions,
			reach = EXCLUDED.reach,
			profile_visits = EXCLUDED.profile_visits,
			follower_count = EXCLUDED.follower_count,
			engagement_rate = EXCLUDED.engagement_rate,
			growth_rate = EXCLUDED.growth_rate,
			updated_at = NOW()
	`

	_, err := r.db.ExecContext(ctx, query,
		insights.AccountID,
		insights.Impressions,
		insights.Reach,
		insights.ProfileVisits,
		insights.FollowerCount,
		insights.EngagementRate,
		insights.GrowthRate,
		insights.MetricDate,
	)

	if err != nil {
		r.logger.Error("Failed to store account insights", zap.Error(err), zap.String("account_id", insights.AccountID))
		return fmt.Errorf("failed to store account insights: %w", err)
	}

	r.logger.Debug("Account insights stored", zap.String("account_id", insights.AccountID))
	return nil
}

// GetAccountInsightsByDate retrieves account insights for a specific date
func (r *instagramInsightsRepository) GetAccountInsightsByDate(ctx context.Context, accountID string, date time.Time) (*models.AccountInsights, error) {
	query := `
		SELECT id, account_id, impressions, reach, profile_visits, follower_count,
			engagement_rate, growth_rate, metric_date, collected_at, updated_at
		FROM instagram_account_insights
		WHERE account_id = $1 AND DATE(metric_date) = DATE($2)
		LIMIT 1
	`

	var insights models.AccountInsights
	err := r.db.QueryRowxContext(ctx, query, accountID, date).StructScan(&insights)
	if err != nil {
		r.logger.Debug("Account insights not found", zap.String("account_id", accountID), zap.String("date", date.String()))
		return nil, fmt.Errorf("account insights not found")
	}

	return &insights, nil
}

// GetAccountInsightsTrend retrieves account insights for the last N days
func (r *instagramInsightsRepository) GetAccountInsightsTrend(ctx context.Context, accountID string, days int) ([]*models.AccountInsights, error) {
	query := `
		SELECT id, account_id, impressions, reach, profile_visits, follower_count,
			engagement_rate, growth_rate, metric_date, collected_at, updated_at
		FROM instagram_account_insights
		WHERE account_id = $1 AND metric_date >= NOW() - INTERVAL '1 day' * $2
		ORDER BY metric_date DESC
	`

	var insights []*models.AccountInsights
	err := r.db.SelectContext(ctx, &insights, query, accountID, days)
	if err != nil {
		r.logger.Error("Failed to get insights trend", zap.Error(err), zap.String("account_id", accountID))
		return nil, fmt.Errorf("failed to get insights trend: %w", err)
	}

	return insights, nil
}
