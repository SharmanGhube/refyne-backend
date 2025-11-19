package auth

import (
	"github.com/google/wire"
	authHandler "github.com/refynehq/refyne-backend/internal/domains/auth/handler"
	authRepo "github.com/refynehq/refyne-backend/internal/domains/auth/repository"
	authService "github.com/refynehq/refyne-backend/internal/domains/auth/services"
)

var ProviderSet = wire.NewSet(
	// Registry
	NewAuthRegistry,

	// Handlers
	authHandler.NewAuthHandler,

	// Services
	authService.NewAuthService,

	// Repositories
	authRepo.NewPasswordResetRepository,
)
