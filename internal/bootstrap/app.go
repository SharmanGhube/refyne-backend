package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/refynehq/refyne-backend/internal/config"
	"github.com/refynehq/refyne-backend/internal/database/migrations"
	authUtils "github.com/refynehq/refyne-backend/internal/domains/auth/utils"
	riverqueue "github.com/refynehq/refyne-backend/internal/shared/river"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type App struct {
	config       *config.Config
	DB           *sqlx.DB
	DBPool       *pgxpool.Pool
	router       *gin.Engine
	server       *http.Server
	logger       *zap.Logger
	riverService *riverqueue.RiverService
	redisClient  *redis.Client
	Version      string
}

func NewApp(
	cfg *config.Config,
	db *sqlx.DB,
	dbPool *pgxpool.Pool,
	router *gin.Engine,
	logger *zap.Logger,
	riverService *riverqueue.RiverService,
	redisClient *redis.Client,
) (*App, error) {
	// Initialize token blacklist manager with Redis
	authUtils.InitTokenBlacklistManager(redisClient)
	logger.Info("Token blacklist manager initialized with Redis")

	// Initialize OTP manager with Redis
	authUtils.InitOTPManager(redisClient)
	logger.Info("OTP manager initialized with Redis")

	app := &App{
		config:       cfg,
		DB:           db,
		DBPool:       dbPool,
		router:       router,
		logger:       logger,
		riverService: riverService,
		redisClient:  redisClient,
		Version:      os.Getenv("REFYNE_VERSION"),
	}

	return app, nil
}

func (a *App) Start(ctx context.Context) error {
	if err := logging.Initialize(); err != nil {
		fmt.Printf("Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}

	defer logging.Close()
	a.logger = logging.GetLogger()
	a.logger.Info("Starting application", zap.String("version", a.Version), zap.String("environment", a.config.Environment))

	// Run Migrations
	if err := a.runMigrations(); err != nil {
		a.logger.Error("Failed to run migrations", zap.Error(err))
		return err
	}

	// Start HTTP server
	addr := fmt.Sprintf(":%s", a.config.Port)
	a.server = &http.Server{
		Addr:         addr,
		Handler:      a.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	a.logger.Info("Starting HTTP server", zap.String("address", addr))

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	a.logger.Info("Server started successfully", zap.String("address", addr))

	// Metrics are exposed at /metrics endpoint for Grafana Cloud scraping
	a.logger.Info("Prometheus metrics available at /metrics endpoint")

	return nil

}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("Stopping application")

	// Stop River service
	if a.riverService != nil {
		a.logger.Info("Stopping River service")
		if err := a.riverService.Stop(); err != nil {
			a.logger.Error("Failed to stop River service", zap.Error(err))
		} else {
			a.logger.Info("River service stopped successfully")
		}
	}

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.Error("Failed to gracefully shutdown HTTP server", zap.Error(err))
		return err
	} else {
		a.logger.Info("HTTP server stopped gracefully")
	}

	// Close database pool
	if a.DBPool != nil {
		a.logger.Info("Closing database pool")
		a.DBPool.Close()
		a.logger.Info("Database pool closed")
	}

	// Close database connection
	if a.DB != nil {
		a.logger.Info("Closing database connection")
		if err := a.DB.Close(); err != nil {
			a.logger.Error("Failed to close database connection", zap.Error(err))
			return err
		}
		a.logger.Info("Database connection closed")
	}

	// Sync logger to ensure all logs are written, but don't close yet
	// The main function will handle the final close
	a.logger.Info("Application components stopped successfully")
	if err := a.logger.Sync(); err != nil {
		// Ignore sync errors on Windows (common issue with file handles)
		// Windows file handles behave differently and can cause sync errors during shutdown
		_ = err // Explicitly ignore the error
	}

	return nil
}

func (a *App) runMigrations() error {
	a.logger.Info("Starting database migration check", zap.String("environment", a.config.Environment), zap.Bool("auto_migrate", a.config.Database.AutoMigrate))

	if a.config.Environment == "production" {
		if !a.config.Database.AutoMigrate {
			a.logger.Info("Skipping migrations in production", zap.Bool("auto_migrate", false))
			return nil
		}

		a.logger.Warn("AUTO_MIGRATE=true in production - applying migrations automatically")

		version, dirty, err := migrations.GetCurrentVersion()
		if err != nil {
			a.logger.Error("Failed to check migration status", zap.Error(err))
			return err
		}

		if dirty {
			a.logger.Error("Database is in dirty state", zap.Uint("version", version))
			return fmt.Errorf("database is in dirty state at version %d", version)
		}

		a.logger.Info("Current migration status", zap.Uint("version", version), zap.Bool("dirty", dirty))
	}

	a.logger.Info("Attempting to run migrations")
	if err := migrations.MigrateUp(); err != nil {
		if err == migrate.ErrNoChange {
			a.logger.Info("No migrations to apply")
		} else {
			a.logger.Error("Migration failed", zap.Error(err))
			if a.config.Environment == "production" {
				return err
			}
		}
	} else {
		a.logger.Info("Migrations applied successfully")
	}

	return nil
}

// GetRouter returns the Gin router instance (useful for testing)
func (a *App) GetRouter() *gin.Engine {
	return a.router
}
