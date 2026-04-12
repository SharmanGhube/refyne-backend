package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/repository"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/services"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// FetchInsightsArgs represents arguments for fetching account insights
type FetchInsightsArgs struct {
	AccountID string `json:"account_id"`
	TrendDays int    `json:"trend_days"` // 0 = current day only, >0 = trend analysis
}

func (FetchInsightsArgs) Kind() string { return "instagram_fetch_insights" }

// FetchInsightsWorker fetches account insights from Instagram API and stores them
type FetchInsightsWorker struct {
	river.WorkerDefaults[FetchInsightsArgs]
	insightsService services.InstagramInsightsService
	insightsRepo    repository.InstagramInsightsRepository
	db              *sqlx.DB
	logger          *zap.Logger
}

// NewFetchInsightsWorker creates a new insights fetching worker
func NewFetchInsightsWorker(
	insightsService services.InstagramInsightsService,
	insightsRepo repository.InstagramInsightsRepository,
	db *sqlx.DB,
) *FetchInsightsWorker {
	return &FetchInsightsWorker{
		insightsService: insightsService,
		insightsRepo:    insightsRepo,
		db:              db,
		logger:          logging.GetJobLogger("FetchInsightsWorker"),
	}
}

// Work fetches insights from Instagram API and stores them in the database
func (w *FetchInsightsWorker) Work(ctx context.Context, job *river.Job[FetchInsightsArgs]) error {
	w.logger.Info("Starting insights fetch",
		zap.String("account_id", job.Args.AccountID),
		zap.Int("trend_days", job.Args.TrendDays),
	)

	// Get account to retrieve access token
	account := &models.InstagramAccount{}
	accountQuery := "SELECT id, user_id, access_token FROM instagram_accounts WHERE id = $1 LIMIT 1"
	err := w.db.QueryRowContext(ctx, accountQuery, job.Args.AccountID).Scan(&account.ID, &account.UserID, &account.AccessToken)
	if err != nil {
		w.logger.Error("Failed to retrieve account access token",
			zap.Error(err),
			zap.String("account_id", job.Args.AccountID),
		)
		return err
	}

	if account.AccessToken == "" {
		w.logger.Error("Account has no access token",
			zap.String("account_id", job.Args.AccountID),
		)
		return fmt.Errorf("no access token for account %s", job.Args.AccountID)
	}

	// Fetch insights
	var insights *models.AccountInsights
	var apiErr error

	if job.Args.TrendDays > 0 {
		// Fetch trend data - for now we'll just fetch current day
		// A more sophisticated implementation would cache daily snapshots
		insights, apiErr = w.insightsService.FetchAccountInsights(ctx, job.Args.AccountID, account.AccessToken)
	} else {
		// Fetch current insights
		insights, apiErr = w.insightsService.FetchAccountInsights(ctx, job.Args.AccountID, account.AccessToken)
	}

	if apiErr != nil {
		w.logger.Error("Failed to fetch insights from Instagram API",
			zap.Error(apiErr),
			zap.String("account_id", job.Args.AccountID),
		)
		return apiErr
	}

	if insights == nil {
		w.logger.Warn("No insights data returned",
			zap.String("account_id", job.Args.AccountID),
		)
		return nil
	}

	// Store insights in database
	if err := w.insightsRepo.StoreAccountInsights(ctx, insights); err != nil {
		w.logger.Error("Failed to store account insights",
			zap.Error(err),
			zap.String("account_id", job.Args.AccountID),
		)
		return err
	}

	// Update account's last_insights_at timestamp
	updateQuery := "UPDATE instagram_accounts SET last_insights_at = $1 WHERE id = $2"
	_, err = w.db.ExecContext(ctx, updateQuery, time.Now(), job.Args.AccountID)
	if err != nil {
		w.logger.Warn("Failed to update last_insights_at", zap.Error(err))
	}

	w.logger.Info("Successfully stored account insights",
		zap.String("account_id", job.Args.AccountID),
		zap.Int("impressions", insights.Impressions),
		zap.Int("reach", insights.Reach),
		zap.Int("follower_count", insights.FollowerCount),
	)

	return nil
}
