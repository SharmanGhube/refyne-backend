package auth

import (
	"github.com/google/wire"
	authHandler "github.com/refynehq/refyne-backend/internal/domain/auth/handler"
	authServices "github.com/refynehq/refyne-backend/internal/domain/auth/services"
)

var ProviderSet = wire.NewSet(
	// Handlers go here
	authHandler.NewAuthHandler,

	// Services go here
	authServices.NewAuthService,

// Repositories go here

// Other dependencies can be added here
)
