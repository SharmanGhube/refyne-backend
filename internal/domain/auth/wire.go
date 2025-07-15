package auth

import (
	"github.com/google/wire"
	authHandler "github.com/refynehq/refyne-backend/internal/domain/auth/handler"
)

var ProviderSet = wire.NewSet(
	// Handlers go here
	authHandler.NewAuthHandler,

// Services go here

// Repositories go here

// Other dependencies can be added here
)
