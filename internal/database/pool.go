package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

func newDatabasePool() (*pgxpool.Pool, error) {
	logger := logging.GetLogger()

	// Build connection string for pgx pool
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	sslMode := os.Getenv("DB_SSL_MODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslMode)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		logger.Error("Failed to parse database config for pool", zap.Error(err))
		return nil, err
	}

	// Configure pool settings based on environment
	env := os.Getenv("APP_ENV")
	if env == "production" {
		config.MaxConns = 20
		config.MinConns = 5
		config.MaxConnLifetime = 15 * time.Minute
		config.MaxConnIdleTime = 5 * time.Minute
	} else {
		// Development settings - more conservative
		config.MaxConns = 10
		config.MinConns = 2
		config.MaxConnLifetime = 1 * time.Hour
		config.MaxConnIdleTime = 10 * time.Minute
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		logger.Error("Failed to create database pool", zap.Error(err))
		return nil, err
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("Failed to ping database pool", zap.Error(err))
		return nil, err
	}

	logger.Info("Database pool initialized successfully")
	return pool, nil
}
