package jobs

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// DailySyncArgs represents arguments for the periodic daily sync job
type DailySyncArgs struct{}

func (DailySyncArgs) Kind() string { return "instagram_daily_sync" }

// DailySyncWorker periodically queues media sync jobs for all active accounts
type DailySyncWorker struct {
	river.WorkerDefaults[DailySyncArgs]
	db          *sqlx.DB
	riverClient *river.Client[any]
	logger      *zap.Logger
}

// NewDailySyncWorker creates a new periodic sync worker
func NewDailySyncWorker(
	db *sqlx.DB,
) *DailySyncWorker {
	return &DailySyncWorker{
		db:     db,
		logger: logging.GetJobLogger("DailySyncWorker"),
	}
}

// SetRiverClient injects the river client after initialization to avoid wire cycles
func (w *DailySyncWorker) SetRiverClient(client *river.Client[any]) {
	w.riverClient = client
}

// Work fetches all active Instagram accounts and queues a sync job for each one
func (w *DailySyncWorker) Work(ctx context.Context, job *river.Job[DailySyncArgs]) error {
	w.logger.Info("Starting periodic daily sync")

	// Get all active account IDs
	query := `
		SELECT id FROM instagram_accounts 
		WHERE deleted_at IS NULL 
		AND token_expires_at > NOW()
	`
	var accountIDs []string
	if err := w.db.SelectContext(ctx, &accountIDs, query); err != nil {
		w.logger.Error("Failed to fetch active accounts for daily sync", zap.Error(err))
		return err
	}

	w.logger.Info("Found active accounts for daily sync", zap.Int("count", len(accountIDs)))

	// Queue a sync job for each account
	var errs int
	for _, accountID := range accountIDs {
		// Enqueue the sync job. 
		// Note: To make this "production grade" and avoid thundering herds, we could add jitter.
		// For now we enqueue them immediately, River will process them up to its concurrency limit (default 100), 
		// so it inherently controls the rate.
		jobArgs := SyncMediaArgs{
			AccountID: accountID,
			SyncType:  "insights",
			Force:     false,
		}

		if _, err := w.riverClient.Insert(ctx, jobArgs, nil); err != nil {
			w.logger.Error("Failed to queue daily sync job", zap.Error(err), zap.String("account_id", accountID))
			errs++
		}
	}

	w.logger.Info("Periodic daily sync complete", 
		zap.Int("queued", len(accountIDs)-errs), 
		zap.Int("failed", errs),
	)

	return nil
}
