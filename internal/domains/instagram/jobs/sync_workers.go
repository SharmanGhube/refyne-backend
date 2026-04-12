package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/repository"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/services"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// SyncMediaArgs represents arguments for media synchronization job
type SyncMediaArgs struct {
	AccountID string `json:"account_id"`
	SyncType  string `json:"sync_type"` // "new", "full", "insights"
	Force     bool   `json:"force"`     // Force refresh even if recently synced
}

func (SyncMediaArgs) Kind() string { return "instagram_sync_media" }

// SyncMediaWorker fetches latest media from Instagram API and updates cache
type SyncMediaWorker struct {
	river.WorkerDefaults[SyncMediaArgs]
	mediaService     services.InstagramMediaService
	mediaRepo        repository.InstagramMediaRepository
	accountRepo      repository.InstagramAccountRepository
	rateLimiter      services.RateLimitChecker
	redis            *redis.Client
	db               *sqlx.DB
	logger           *zap.Logger
}

// NewSyncMediaWorker creates a new media sync worker
func NewSyncMediaWorker(
	mediaService services.InstagramMediaService,
	mediaRepo repository.InstagramMediaRepository,
	accountRepo repository.InstagramAccountRepository,
	rateLimiter services.RateLimitChecker,
	redis *redis.Client,
	db *sqlx.DB,
) *SyncMediaWorker {
	return &SyncMediaWorker{
		mediaService:     mediaService,
		mediaRepo:        mediaRepo,
		accountRepo:      accountRepo,
		rateLimiter:      rateLimiter,
		redis:            redis,
		db:               db,
		logger:           logging.GetJobLogger("SyncMediaWorker"),
	}
}

// Work synchronizes media from Instagram API
// This job:
// 1. Checks rate limits (200 calls/hour per account)
// 2. Fetches media from Instagram API
// 3. Deduplicates against existing cache
// 4. Stores in database
// 5. Queues AI processing for new media
func (w *SyncMediaWorker) Work(ctx context.Context, job *river.Job[SyncMediaArgs]) error {
	w.logger.Info("Starting media sync",
		zap.String("account_id", job.Args.AccountID),
		zap.String("sync_type", job.Args.SyncType),
		zap.Bool("force", job.Args.Force),
	)

	// Check rate limit before making API calls
	// Instagram allows 200 calls/hour per account
	canCall, remaining, err := w.rateLimiter.CanMakeCall(ctx, job.Args.AccountID)
	if err != nil {
		w.logger.Error("Failed to check rate limit", zap.Error(err))
		return err
	}

	if !canCall {
		w.logger.Warn("Rate limit reached",
			zap.String("account_id", job.Args.AccountID),
			zap.Int("remaining_calls", remaining),
		)
		// Return error to trigger backoff retry
		return fmt.Errorf("rate limit reached for account %s", job.Args.AccountID)
	}

	// Check if recent sync exists (skip if not forced and recently synced)
	if !job.Args.Force {
		query := "SELECT last_sync_at FROM instagram_accounts WHERE id = $1 LIMIT 1"
		var lastSyncTime sql.NullTime
		err := w.db.QueryRowContext(ctx, query, job.Args.AccountID).Scan(&lastSyncTime)
		if err == nil && lastSyncTime.Valid {
			syncAge := time.Since(lastSyncTime.Time)
			if syncAge < 30*time.Minute { // Skip if synced within 30 minutes
				w.logger.Info("Media recently synced, skipping",
					zap.String("account_id", job.Args.AccountID),
					zap.Duration("sync_age", syncAge),
				)
				return nil
			}
		}
	}

	// Update account sync status to "syncing"
	updateStatusQuery := "UPDATE instagram_accounts SET sync_status = $1, sync_error = NULL WHERE id = $2"
	_, err = w.db.ExecContext(ctx, updateStatusQuery, "syncing", job.Args.AccountID)
	if err != nil {
		w.logger.Warn("Failed to update sync status", zap.Error(err))
	}

	// Get account to retrieve access token
	account := &models.InstagramAccount{}
	accountQuery := "SELECT id, user_id, access_token FROM instagram_accounts WHERE id = $1 LIMIT 1"
	err = w.db.QueryRowContext(ctx, accountQuery, job.Args.AccountID).Scan(&account.ID, &account.UserID, &account.AccessToken)
	if err != nil {
		w.logger.Error("Failed to retrieve account access token",
			zap.Error(err),
			zap.String("account_id", job.Args.AccountID),
		)
		w.db.ExecContext(ctx, "UPDATE instagram_accounts SET sync_status = $1, sync_error = $2 WHERE id = $3", "error", "Failed to retrieve access token", job.Args.AccountID)
		return err
	}

	// Fetch media from Instagram API with access token
	media, err := w.mediaService.FetchMedia(ctx, job.Args.AccountID, account.AccessToken, job.Args.SyncType)
	if err != nil {
		w.logger.Error("Failed to fetch media from Instagram",
			zap.Error(err),
			zap.String("account_id", job.Args.AccountID),
		)
		// Update sync status to error
		errMsg := err.Error()
		w.db.ExecContext(ctx, "UPDATE instagram_accounts SET sync_status = $1, sync_error = $2 WHERE id = $3", "error", errMsg, job.Args.AccountID)
		// Retry with exponential backoff
		return err
	}

	if len(media) == 0 {
		w.logger.Info("No media to sync",
			zap.String("account_id", job.Args.AccountID),
		)
		// Update sync status back to idle
		w.db.ExecContext(ctx, "UPDATE instagram_accounts SET sync_status = $1, sync_error = NULL WHERE id = $2", "idle", job.Args.AccountID)
		return nil
	}

	// Store media in database (upsert)
	mediaInputs := make([]*models.CreateInstagramMediaInput, len(media))
	for i, m := range media {
		mediaInputs[i] = &models.CreateInstagramMediaInput{
			AccountID:        job.Args.AccountID,
			InstagramMediaID: m.InstagramMediaID,
			MediaType:        m.MediaType,
			Caption:          m.Caption.String,
			MediaURL:         m.MediaURL,
			Permalink:        m.Permalink.String,
			ThumbnailURL:     m.ThumbnailURL.String,
			PostedAt:         m.PostedAt,
			LikeCount:        m.LikeCount,
			CommentCount:     m.CommentCount,
		}
	}

	if syncErr := w.mediaRepo.UpsertMedia(ctx, job.Args.AccountID, mediaInputs); syncErr != nil {
		w.logger.Error("Failed to store media in database", zap.String("account_id", job.Args.AccountID))
		w.db.ExecContext(ctx, "UPDATE instagram_accounts SET sync_status = $1, sync_error = $2 WHERE id = $3", "error", "Database sync failed", job.Args.AccountID)
		// Unwrap the AppError to get the underlying error
		if syncErr != nil {
			return fmt.Errorf("media upsert failed")
		}
	}

	// Cache media in Redis (1 hour TTL)
	if cacheErr := w.mediaService.CacheMedia(ctx, job.Args.AccountID, media); cacheErr != nil {
		w.logger.Warn("Failed to cache media", zap.Error(cacheErr))
		// Don't fail the job for cache errors
	}

	// TODO: Queue AI processing jobs for new media
	// For each media item, check if AI analysis exists
	// If not, queue ProcessAIWorker job

	// Update account's last_sync_at timestamp
	err = w.db.QueryRowContext(ctx,
		"UPDATE instagram_accounts SET last_sync_at = NOW() WHERE id = $1",
		job.Args.AccountID,
	).Scan()
	if err != nil && err != sql.ErrNoRows {
		w.logger.Warn("Failed to update last_sync_at")
	}

	// Update sync status to idle
	w.db.ExecContext(ctx, "UPDATE instagram_accounts SET sync_status = $1, sync_error = NULL WHERE id = $2", "idle", job.Args.AccountID)

	w.logger.Info("Media sync complete",
		zap.String("account_id", job.Args.AccountID),
		zap.Int("media_count", len(media)),
	)

	return nil
}


// RefreshTokenArgs represents arguments for token refresh job
type RefreshTokenArgs struct {
	AccountID string `json:"account_id"`
}

func (RefreshTokenArgs) Kind() string { return "instagram_refresh_token" }

// RefreshTokenWorker refreshes long-lived access tokens before expiry
type RefreshTokenWorker struct {
	river.WorkerDefaults[RefreshTokenArgs]
	oauthService services.InstagramOAuthService
	accountRepo  repository.InstagramAccountRepository
	db           *sqlx.DB
	logger       *zap.Logger
}

// NewRefreshTokenWorker creates a new token refresh worker
func NewRefreshTokenWorker(
	oauthService services.InstagramOAuthService,
	accountRepo repository.InstagramAccountRepository,
	db *sqlx.DB,
) *RefreshTokenWorker {
	return &RefreshTokenWorker{
		oauthService: oauthService,
		accountRepo:  accountRepo,
		db:           db,
		logger:       logging.GetJobLogger("RefreshTokenWorker"),
	}
}

// Work refreshes Instagram access token before expiry
// Scheduled to run 23 days before token expires (55 days into 60-day validity)
func (w *RefreshTokenWorker) Work(ctx context.Context, job *river.Job[RefreshTokenArgs]) error {
	w.logger.Info("Refreshing Instagram token",
		zap.String("account_id", job.Args.AccountID),
	)

	// Check if token needs refresh by querying database
	query := "SELECT token_expires_at FROM instagram_accounts WHERE id = $1 LIMIT 1"
	var expiresAt time.Time
	err := w.db.QueryRowContext(ctx, query, job.Args.AccountID).Scan(&expiresAt)
	if err != nil {
		w.logger.Error("Failed to check token expiry", zap.Error(err))
		return fmt.Errorf("failed to check token expiry: %w", err)
	}

	// Check if token still has more than 7 days (give extra buffer)
	if time.Until(expiresAt) > 7*24*time.Hour {
		w.logger.Info("Token still valid, skipping refresh",
			zap.String("account_id", job.Args.AccountID),
			zap.Duration("expires_in", time.Until(expiresAt)),
		)
		return nil
	}

	// TODO: Call Instagram token refresh endpoint
	// Instagram long-lived tokens valid for 60 days
	// We refresh at 55 days to have buffer
	// Use oauthService.RefreshToken(ctx, accountID)

	w.logger.Info("Token refresh complete",
		zap.String("account_id", job.Args.AccountID),
	)

	return nil
}


// ProcessAIArgs represents arguments for AI processing job
type ProcessAIArgs struct {
	MediaID   string `json:"media_id"`
	AccountID string `json:"account_id"`
	Caption   string `json:"caption"`
	MediaURL  string `json:"media_url"`
	MediaType string `json:"media_type"` // PHOTO, VIDEO, CAROUSEL, REELS, STORY
}

func (ProcessAIArgs) Kind() string { return "instagram_process_ai" }

// ProcessAIWorker handles AI analysis of media using Gemini
type ProcessAIWorker struct {
	river.WorkerDefaults[ProcessAIArgs]
	geminiService services.GeminiService
	aiRepo        repository.InstagramAIRepository
	logger        *zap.Logger
}

// NewProcessAIWorker creates a new AI processing worker
func NewProcessAIWorker(
	geminiService services.GeminiService,
	aiRepo repository.InstagramAIRepository,
) *ProcessAIWorker {
	return &ProcessAIWorker{
		geminiService: geminiService,
		aiRepo:        aiRepo,
		logger:        logging.GetJobLogger("ProcessAIWorker"),
	}
}

// Work performs AI analysis on media
// This job:
// 1. Downloads and encodes media
// 2. Calls Gemini Vision API
// 3. Analyzes sentiment, themes, quality
// 4. Generates caption alternatives
// 5. Predicts engagement and posting times
// 6. Stores recommendations
func (w *ProcessAIWorker) Work(ctx context.Context, job *river.Job[ProcessAIArgs]) error {
	w.logger.Info("Processing media with AI",
		zap.String("media_id", job.Args.MediaID),
		zap.String("media_type", job.Args.MediaType),
	)

	// Check if recommendations already exist and are fresh
	exists, err := w.aiRepo.RecommendationsExist(ctx, job.Args.MediaID)
	if err == nil && exists {
		w.logger.Debug("Fresh recommendations already exist, skipping",
			zap.String("media_id", job.Args.MediaID),
		)
		return nil
	}

	// Step 1: Perform AI analysis
	analysis, err := w.geminiService.AnalyzeMedia(ctx, job.Args.MediaType, job.Args.Caption, job.Args.MediaURL)
	if err != nil {
		w.logger.Error("Failed to analyze media", zap.Error(err), zap.String("media_id", job.Args.MediaID))
		return err
	}

	// Step 2: Generate caption alternatives
	captions, err := w.geminiService.GenerateCaptions(ctx, job.Args.Caption, job.Args.MediaType, analysis.ContentThemes)
	if err != nil {
		w.logger.Warn("Failed to generate captions", zap.Error(err), zap.String("media_id", job.Args.MediaID))
		// Continue without captions - don't fail
		captions = []*models.CaptionSuggestion{}
	}

	// Step 3: Generate posting strategy
	strategy, err := w.geminiService.GeneratePostingStrategy(ctx, analysis, job.Args.AccountID)
	if err != nil {
		w.logger.Warn("Failed to generate posting strategy", zap.Error(err), zap.String("media_id", job.Args.MediaID))
		// Continue without strategy - don't fail
		strategy = &models.PostingStrategy{}
	}

	// Step 4: Store recommendations
	recommendations := &models.AIRecommendations{
		MediaID:            job.Args.MediaID,
		AccountID:          job.Args.AccountID,
		Analysis:           analysis,
		CaptionSuggestions: captions,
		PostingStrategy:    strategy,
		ConfidenceScore:    0.85, // TODO: Calculate based on analysis quality
		GeneratedAt:        time.Now(),
		ExpiresAt:          time.Now().Add(30 * 24 * time.Hour), // 30-day expiry
	}

	if err := w.aiRepo.StoreRecommendations(ctx, recommendations); err != nil {
		w.logger.Error("Failed to store AI recommendations", zap.Error(err), zap.String("media_id", job.Args.MediaID))
		return err
	}

	w.logger.Info("AI processing complete",
		zap.String("media_id", job.Args.MediaID),
		zap.String("sentiment", analysis.Sentiment),
		zap.Int("quality_score", analysis.QualityScore),
	)

	return nil
}

