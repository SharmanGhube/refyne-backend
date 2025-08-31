package user

import (
	"github.com/google/wire"
	userRepo "github.com/refynehq/refyne-backend/internal/domains/user/core/repository"
)

var ProviderSet = wire.NewSet(
	// Registry
	NewUserRegistry,

	// Repository
	userRepo.NewCoreUserRepository,

	// Handlers

	// Services

)
