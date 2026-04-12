package riverqueue

import (
	"context"
	"time"
	"unsafe"

	"github.com/jackc/pgx/v5/pgxpool"
	emailJobs "github.com/refynehq/refyne-backend/internal/domains/email/jobs"
	instagramJobs "github.com/refynehq/refyne-backend/internal/domains/instagram/jobs"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"go.uber.org/zap"
)

type RiverService struct {
	client *river.Client[any]
	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger
}
type WorkerDependancies struct {
	// User

	// Email
	EmailWorker *emailJobs.EmailWorker

	// Instagram
	InstagramWebhookWorker *instagramJobs.InstagramWebhookWorker
	SyncMediaWorker        *instagramJobs.SyncMediaWorker
	FetchInsightsWorker    *instagramJobs.FetchInsightsWorker
	RefreshTokenWorker     *instagramJobs.RefreshTokenWorker
	ProcessAIWorker        *instagramJobs.ProcessAIWorker
}

// I have no idea what the fuck is going on here
func NewRiverService(dbPool *pgxpool.Pool, deps *WorkerDependancies) (*RiverService, error) {
	logger := logging.GetServiceLogger("riverqueue")
	logger.Info("Initializing River Service")

	// Validate dependencies
	if dbPool == nil {
		return nil, NewInvalidDependencyError(nil, "Database pool is required for River Service")
	}

	// Check if river schema exists
	var exists bool
	err := dbPool.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'river_job')").Scan(&exists)
	if err != nil {
		logger.Warn("Failed to check if River schema exists, proceeding with migration", zap.Error(err))
		exists = false
	}

	// Only run migrations if schema does not exist
	if !exists {
		logger.Info("Running River Migrations")

		migrator, err := rivermigrate.New(riverpgxv5.New(dbPool), nil)
		if err != nil {
			logger.Error("Failed to create River migrator", zap.Error(err))
			return nil, NewRiverMigratorCreationError(nil, "Failed to create River migrator")
		}

		_, err = migrator.Migrate(context.Background(), rivermigrate.DirectionUp, nil)
		if err != nil {
			logger.Error("River migration failed", zap.Error(err))
			return nil, NewRiverMigrationError(nil, "River migration failed")
		}

		logger.Info("River migrations completed successfully")

	} else {
		logger.Info("River schema already exists, skipping migrations")
	}

	workers := river.NewWorkers()
	// Register workers
	if deps.EmailWorker != nil {
		river.AddWorker(workers, deps.EmailWorker)
		logger.Info("Registered Email Worker")
	}

	// Instagram workers
	if deps.InstagramWebhookWorker != nil {
		river.AddWorker(workers, deps.InstagramWebhookWorker)
		logger.Info("Registered Instagram Webhook Worker")
	}
	if deps.SyncMediaWorker != nil {
		river.AddWorker(workers, deps.SyncMediaWorker)
		logger.Info("Registered Sync Media Worker")
	}
	if deps.FetchInsightsWorker != nil {
		river.AddWorker(workers, deps.FetchInsightsWorker)
		logger.Info("Registered Fetch Insights Worker")
	}
	if deps.RefreshTokenWorker != nil {
		river.AddWorker(workers, deps.RefreshTokenWorker)
		logger.Info("Registered Refresh Token Worker")
	}
	if deps.ProcessAIWorker != nil {
		river.AddWorker(workers, deps.ProcessAIWorker)
		logger.Info("Registered Process AI Worker")
	}

	logger.Info("Registered River Workers")

	// Periodic Jobs - scheduled via external scheduler or manual queueing
	// These jobs are typically triggered by:
	// - SyncMediaWorker: Manual API call or external scheduler (30min interval)
	// - FetchInsightsWorker: Manual API call or external scheduler (daily)
	// - RefreshTokenWorker: Manual API call or external scheduler (50-day interval)
	periodicJobs := []*river.PeriodicJob{}

	// TODO: In production, use external scheduler (cron, AWS EventBridge, etc.)
	// or implement internal scheduler with go-cron package

	// Create River client
	riverClient, err := river.NewClient(
		riverpgxv5.New(dbPool),
		&river.Config{
			Workers:      workers,
			Queues:       DefaultQueueConfig().GetQueues(),
			PeriodicJobs: periodicJobs,
		},
	)
	if err != nil {
		logger.Error("Failed to create River client", zap.Error(err))
		return nil, NewRiverClientCreationError(nil, "Failed to create River client")
	}

	ctx, cancel := context.WithCancel(context.Background())
	service := &RiverService{
		client: (*river.Client[any])(unsafe.Pointer(riverClient)),
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}

	logger.Info("River service initialized successfully")
	return service, nil
}

func (s *RiverService) Start() *errors.AppError {
	s.logger.Info("Starting River Service")

	if err := s.client.Start(s.ctx); err != nil {
		s.logger.Error("Failed to start river service",
			zap.Error(err),
		)
		return NewRiverStartError(nil, "Failed to start river service", err)
	}

	s.logger.Info("River Service started successfully")
	return nil
}

func (s *RiverService) Stop() *errors.AppError {
	s.logger.Info("Stopping river queue service")

	// Cancel the service context to signal shutdown
	s.cancel()

	// Create a new context with timeout for the stop operation
	// This prevents the "context canceled" error since we're not using the canceled context
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := s.client.Stop(stopCtx); err != nil {
		s.logger.Error("Failed to stop river service",
			zap.Error(err))

		return NewRiverStopError(nil, "Failed to stop river service", err)
	}

	s.logger.Info("River Service stopped successfully")
	return nil
}

func (s *RiverService) GetClient() *river.Client[any] {
	return s.client
}
