package auth

import (
	"github.com/google/wire"
	authHandler "github.com/refynehq/refyne-backend/internal/domains/auth/handler"
	authService "github.com/refynehq/refyne-backend/internal/domains/auth/services"
)

var ProviderSet = wire.NewSet(
	// Registry
	NewAuthRegistry,

	// Handlers
	authHandler.NewAuthHandler,

	// Services
	authService.NewAuthService,
)
