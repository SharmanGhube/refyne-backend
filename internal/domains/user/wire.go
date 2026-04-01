package user

import (
	"github.com/google/wire"
	userRepo "github.com/refynehq/refyne-backend/internal/domains/user/core/repository"
	userHandler "github.com/refynehq/refyne-backend/internal/domains/user/handler"
	userServices "github.com/refynehq/refyne-backend/internal/domains/user/services"
)

var ProviderSet = wire.NewSet(
	// Registry
	NewUserRegistry,

	// Repository
	userRepo.NewCoreUserRepository,

	// Settings Repository
	userServices.NewSettingsRepository,

	// Services
	userServices.NewUserService,

	// Handlers
	userHandler.NewUserHandler,
)
