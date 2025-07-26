package auth

import (
	"github.com/google/wire"
	authHandler "github.com/refynehq/refyne-backend/internal/domain/auth/handler"
	authService "github.com/refynehq/refyne-backend/internal/domain/auth/service"
)

var ProviderSet = wire.NewSet(
	// Registry goes here
	NewAuthRegistry,

	// Handlers go here
	authHandler.NewAuthHandler,

	// Services go here
	authService.NewAuthService,
)
