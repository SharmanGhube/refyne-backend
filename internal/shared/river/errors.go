package riverqueue

import (
	"github.com/gin-gonic/gin"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// Error codes
const (
	CodeInvalidDependency           = "RIVER_INVALID_DEPENDENCY"
	CodeRiverMigratorCreationFailed = "RIVER_MIGRATOR_CREATION_FAILED"
	CodeRiverMigrationFailed        = "RIVER_MIGRATION_FAILED"
	CodeRiverClientCreationFailed   = "RIVER_CLIENT_CREATION_FAILED"
	CodeRiverStartFailed            = "RIVER_START_FAILED"
	CodeRiverStopFailed             = "RIVER_STOP_FAILED"
)

// Error Factory Functions
func NewInvalidDependencyError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeInvalidDependency,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"riverqueue",
	)
}

func NewRiverMigratorCreationError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeRiverMigratorCreationFailed,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"riverqueue",
	)
}

func NewRiverMigrationError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeRiverMigrationFailed,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"riverqueue",
	)
}

func NewRiverClientCreationError(c *gin.Context, message string) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeRiverClientCreationFailed,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"riverqueue",
	)
}

func NewRiverStartError(c *gin.Context, message string, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeRiverStartFailed,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"riverqueue",
	).WithContext("error", err)
}

func NewRiverStopError(c *gin.Context, message string, err error) *errors.AppError {
	return errors.NewAppError(
		c,
		CodeRiverStopFailed,
		message,
		errors.ErrorTypeInternal,
		errors.SeverityHigh,
		"riverqueue",
	).WithContext("error", err)
}
