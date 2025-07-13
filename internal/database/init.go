package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

func initDB() (*sqlx.DB, error) {
	logger := logging.GetLogger()

	logger.Info("Initializing database connection")

	// Get database connection parameters
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSL_MODE")

	// Build connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, database, sslMode)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	env := os.Getenv("APP_ENV")
	if env == "production" {
		// Calculate optimal DB connection pool settings based on application needs
		maxConnections := 20 // Default for production
		if val := os.Getenv("DB_MAX_CONNECTIONS"); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
				maxConnections = parsed
			}
		}

		// Set higher idle connections to reduce connection setup overhead
		maxIdleConnections := maxConnections / 2
		if val := os.Getenv("DB_MAX_IDLE_CONNECTIONS"); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
				maxIdleConnections = parsed
			}
		}

		// Reduce connection lifetime to prevent stale connections
		connMaxLifetime := 15 * time.Minute
		if val := os.Getenv("DB_CONN_MAX_LIFETIME"); val != "" {
			if parsed, err := time.ParseDuration(val); err == nil && parsed > 0 {
				connMaxLifetime = parsed
			}
		}

		// Add connection max idle time
		connMaxIdleTime := 5 * time.Minute
		if val := os.Getenv("DB_CONN_MAX_IDLE_TIME"); val != "" {
			if parsed, err := time.ParseDuration(val); err == nil && parsed > 0 {
				connMaxIdleTime = parsed
			}
		}

		db.SetMaxOpenConns(maxConnections)
		db.SetMaxIdleConns(maxIdleConnections)
		db.SetConnMaxLifetime(connMaxLifetime)
		db.SetConnMaxIdleTime(connMaxIdleTime)

		logger.Info("Database connection pool settings applied",
			zap.Int("max_open_conns", maxConnections),
			zap.Int("max_idle_conns", maxIdleConnections),
			zap.Duration("conn_max_lifetime", connMaxLifetime),
			zap.Duration("conn_max_idle_time", connMaxIdleTime),
		)
	} else {
		// Development settings - more conservative
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(time.Hour)
		db.SetConnMaxIdleTime(10 * time.Minute)
	}

	return db, nil
}
