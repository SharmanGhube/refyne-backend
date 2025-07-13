package migrations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// MigrateUp applies all available migrations
func MigrateUp() error {

	logger := logging.GetServiceLogger("migrations")

	dsn := createConnectionString()
	migrationsPath := getMigrationsPath()

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to apply migrations", zap.Error(err))
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	logger.Info("Migrations applied successfully")
	return nil
}

// MigrateDown rolls back the last migration (one step)
func MigrateDown() error {

	logger := logging.GetServiceLogger("migrations")

	dsn := createConnectionString()
	migrationsPath := getMigrationsPath()

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to rollback migration", zap.Error(err))
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	logger.Info("Migration rolled back successfully")
	return nil
}

// MigrateDownAll rolls back all migrations
func MigrateDownAll() error {

	logger := logging.GetServiceLogger("migrations")

	dsn := createConnectionString()
	migrationsPath := getMigrationsPath()

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to rollback all migrations", zap.Error(err))
		return fmt.Errorf("failed to rollback all migrations: %w", err)
	}

	logger.Info("All migrations rolled back successfully")
	return nil
}

// MigrateToVersion migrates to a specific version
func MigrateToVersion(version uint) error {

	logger := logging.GetServiceLogger("migrations")

	dsn := createConnectionString()
	migrationsPath := getMigrationsPath()

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Migrate(version); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to migrate to version", zap.Uint("version", version), zap.Error(err))
		return fmt.Errorf("failed to migrate to version %d: %w", version, err)
	}

	logger.Info("Migrated to version successfully", zap.Uint("version", version))
	return nil
}

// MigrateSteps migrates up (+) or down (-) by specific number of steps
func MigrateSteps(steps int) error {
	logger := logging.GetServiceLogger("migrations")

	dsn := createConnectionString()
	migrationsPath := getMigrationsPath()

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Steps(steps); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to migrate steps", zap.Int("steps", steps), zap.Error(err))
		return fmt.Errorf("failed to migrate %d steps: %w", steps, err)
	}

	logger.Info("Migrated steps successfully", zap.Int("steps", steps))
	return nil
}

// GetCurrentVersion returns the current migration version and dirty state
func GetCurrentVersion() (uint, bool, error) {
	logger := logging.GetServiceLogger("migrations")

	dsn := createConnectionString()
	migrationsPath := getMigrationsPath()

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return 0, false, fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			logger.Info("No migrations applied yet")
			return 0, false, nil // No migrations applied yet
		}
		logger.Error("Failed to get migration version", zap.Error(err))
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}

	logger.Info("Current migration version", zap.Uint("version", version), zap.Bool("dirty", dirty))
	return version, dirty, nil
}

// Force sets the migration version without running migrations (recovery only)
func Force(version int) error {
	logger := logging.GetServiceLogger("migrations")

	dsn := createConnectionString()
	migrationsPath := getMigrationsPath()

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		logger.Error("Failed to force migration version", zap.Int("version", version), zap.Error(err))
		return fmt.Errorf("failed to force migration version %d: %w", version, err)
	}

	logger.Info("Forced migration to version", zap.Int("version", version))
	return nil
}

// Drop removes all tables and the schema_migrations table (DANGEROUS)
func Drop() error {
	logger := logging.GetServiceLogger("migrations")

	dsn := createConnectionString()
	migrationsPath := getMigrationsPath()

	m, err := migrate.New(
		"file://"+migrationsPath,
		dsn,
	)
	if err != nil {
		logger.Error("Failed to create migration instance", zap.Error(err))
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Drop(); err != nil {
		logger.Error("Failed to drop database", zap.Error(err))
		return fmt.Errorf("failed to drop database: %w", err)
	}

	logger.Info("Database dropped successfully")
	return nil
}

// createConnectionString builds the database connection string from environment variables
func createConnectionString() string {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbSSLMode := os.Getenv("DB_SSL_MODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSLMode)
}

// getMigrationsPath returns the path to the migrations directory
func getMigrationsPath() string {
	logger := logging.GetServiceLogger("migrations")

	// Get the working directory
	workDir, err := os.Getwd()
	if err != nil {
		logger.Warn("Could not get working directory, using fallback", zap.Error(err))
		return "migrations" // Fallback
	}

	// Convert Windows path to proper file:// URL format
	path := filepath.Join(workDir, "/internal/database/migrations")

	// Replace backslashes with forward slashes
	path = filepath.ToSlash(path)

	return path
}
