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
	"github.com/refynehq/refyne-backend/internal/config"
	"github.com/refynehq/refyne-backend/internal/database/migrations"
	riverqueue "github.com/refynehq/refyne-backend/internal/shared/river"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type App struct {
	config       *config.Config
	db           *sqlx.DB
	dbPool       *pgxpool.Pool
	router       *gin.Engine
	server       *http.Server
	logger       *zap.Logger
	riverService *riverqueue.RiverService

	version string
}

func NewApp(cfg *config.Config, db *sqlx.DB, dbPool *pgxpool.Pool, router *gin.Engine, riverService *riverqueue.RiverService) *App {
	return &App{
		config:       cfg,
		db:           db,
		dbPool:       dbPool,
		router:       router,
		riverService: riverService,
		version:      os.Getenv("APP_VERSION"),
	}
}

func (a *App) Start(ctx context.Context) error {
	if err := logging.Initialize(); err != nil {
		fmt.Printf("Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}

	defer logging.Close()
	a.logger = logging.GetLogger()
	a.logger.Info("Starting application", zap.String("version", a.version), zap.String("environment", a.config.Environment))

	// Run Migrations
	if err := a.runMigrations(); err != nil {
		a.logger.Error("Failed to run migrations", zap.Error(err))
		return err
	}

	// Start river service later
	// if err := a.riverService.Start(); err != nil {
	// 	a.logger.Error("Failed to start River service", zap.Error(err))
	// 	return fmt.Errorf("failed to start River service: %w", err)
	// }
	// a.logger.Info("River service started successfully")

	// Start prometheus service

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
	return nil

}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("Stopping application")

	// Stop River service first
	if a.riverService != nil {
		a.logger.Info("Stopping River service")
		if err := a.riverService.Stop(); err != nil {
			a.logger.Error("Failed to stop River service", zap.Error(err))
			// Continue with shutdown process even if River service fails to stop gracefully
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
	if a.dbPool != nil {
		a.logger.Info("Closing database pool")
		a.dbPool.Close()
		a.logger.Info("Database pool closed")
	}

	// Close database connection
	if a.db != nil {
		a.logger.Info("Closing database connection")
		if err := a.db.Close(); err != nil {
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
	a.logger.Info("Starting database migration check")

	if a.config.Environment == "production" {
		if !a.config.AutoMigrate {
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

		a.logger.Info("Current migration status", zap.Uint("version", version))
	}

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
