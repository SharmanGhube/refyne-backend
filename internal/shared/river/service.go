package riverqueue

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"go.uber.org/zap"
)

type RiverService struct {
	client *river.Client[pgx.Tx]
	ctx    context.Context
	cancel context.CancelFunc
	logger *zap.Logger
}
type WorkerDependancies struct {
	// User

	// Instagram
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
	// river.AddWorker(workers, emailWorkers.NewEmailVerificationWorker())

	logger.Info("Registered River Workers")

	// Periodic Jobs
	periodicJobs := []*river.PeriodicJob{}

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
		client: riverClient,
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

func (s *RiverService) GetClient() *river.Client[pgx.Tx] {
	return s.client
}
