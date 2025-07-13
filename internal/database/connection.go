package database

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

func NewDatabase() (*sqlx.DB, error) {
	logger := logging.GetLogger()

	db, err := initDB() // Use private function
	if err != nil {
		logger.Error("Failed to initialize database", zap.Error(err))
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, err
	}

	logger.Info("Database connection established successfully")
	return db, nil
}

func NewPool() (*pgxpool.Pool, error) {
	logger := logging.GetLogger()

	pool, err := newDatabasePool() // Use private function
	if err != nil {
		logger.Error("Failed to initialize database pool", zap.Error(err))
		return nil, err
	}

	logger.Info("Database pool established successfully")
	return pool, nil
}
