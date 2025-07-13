package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// This repository will have Core CRUD operations for the user domain.
// I love torturing myself, gives me the high without the need for drugs.

type CoreUserRepository interface {
	// Core CRUD operations
	// CreateUser() *errors.AppError
}

type coreUserRepository struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

func NewCoreUserRepository(db *sqlx.DB) CoreUserRepository {
	return &coreUserRepository{
		name:   "CoreUserRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("CoreUserRepository"),
	}
}
