package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// QueryExecutor provides secure query execution with timeouts and logging
type QueryExecutor struct {
	db      *sqlx.DB
	logger  *zap.Logger
	timeout time.Duration
}

// NewQueryExecutor creates a new query executor with default timeout
func NewQueryExecutor(db *sqlx.DB, timeout time.Duration) *QueryExecutor {
	if timeout == 0 {
		timeout = 30 * time.Second // Default 30s query timeout
	}

	return &QueryExecutor{
		db:      db,
		logger:  logging.GetComponentLogger("database.executor"),
		timeout: timeout,
	}
}

// ExecContext executes a query with context timeout
// Use this for INSERT, UPDATE, DELETE operations
func (qe *QueryExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	result, err := qe.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		qe.logger.Error("Query execution failed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	// Log slow queries (> 1 second)
	if duration > time.Second {
		qe.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
		)
	}

	return result, nil
}

// QueryContext executes a SELECT query with context timeout
func (qe *QueryExecutor) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	rows, err := qe.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		qe.logger.Error("Query failed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Log slow queries
	if duration > time.Second {
		qe.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
		)
	}

	return rows, nil
}

// QueryRowContext executes a query that returns a single row
func (qe *QueryExecutor) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	row := qe.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	// Log slow queries
	if duration > time.Second {
		qe.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
		)
	}

	return row
}

// GetContext is a helper for SELECT queries that scan into a struct
func (qe *QueryExecutor) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	err := qe.db.GetContext(ctx, dest, query, args...)
	duration := time.Since(start)

	if err != nil {
		if err != sql.ErrNoRows {
			qe.logger.Error("Get query failed",
				zap.String("query", query),
				zap.Duration("duration", duration),
				zap.Error(err),
			)
		}
		return err
	}

	// Log slow queries
	if duration > time.Second {
		qe.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
		)
	}

	return nil
}

// SelectContext is a helper for SELECT queries that scan into a slice
func (qe *QueryExecutor) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	err := qe.db.SelectContext(ctx, dest, query, args...)
	duration := time.Since(start)

	if err != nil {
		qe.logger.Error("Select query failed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return fmt.Errorf("select query failed: %w", err)
	}

	// Log slow queries
	if duration > time.Second {
		qe.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
		)
	}

	return nil
}

// BeginTxContext starts a transaction with context timeout
func (qe *QueryExecutor) BeginTxContext(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	tx, err := qe.db.BeginTx(ctx, opts)
	if err != nil {
		qe.logger.Error("Failed to begin transaction", zap.Error(err))
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, nil
}

// PrepareContext prepares a statement with context timeout
// Prepared statements prevent SQL injection at the database level
func (qe *QueryExecutor) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	stmt, err := qe.db.PrepareContext(ctx, query)
	if err != nil {
		qe.logger.Error("Failed to prepare statement",
			zap.String("query", query),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	return stmt, nil
}

// NamedExecContext executes a named query with context timeout
func (qe *QueryExecutor) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	result, err := qe.db.NamedExecContext(ctx, query, arg)
	duration := time.Since(start)

	if err != nil {
		qe.logger.Error("Named exec failed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, fmt.Errorf("named exec failed: %w", err)
	}

	// Log slow queries
	if duration > time.Second {
		qe.logger.Warn("Slow named query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
		)
	}

	return result, nil
}

// NamedQueryContext executes a named SELECT query with context timeout
func (qe *QueryExecutor) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, qe.timeout)
	defer cancel()

	start := time.Now()
	rows, err := qe.db.NamedQueryContext(ctx, query, arg)
	duration := time.Since(start)

	if err != nil {
		qe.logger.Error("Named query failed",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, fmt.Errorf("named query failed: %w", err)
	}

	// Log slow queries
	if duration > time.Second {
		qe.logger.Warn("Slow named query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
		)
	}

	return rows, nil
}

// SetTimeout allows changing the default query timeout
func (qe *QueryExecutor) SetTimeout(timeout time.Duration) {
	qe.timeout = timeout
}

// GetDB returns the underlying database connection
// Use with caution - prefer using the context methods
func (qe *QueryExecutor) GetDB() *sqlx.DB {
	return qe.db
}
