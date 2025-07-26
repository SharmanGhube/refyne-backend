package user

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/domain/user/account"
	user "github.com/refynehq/refyne-backend/internal/domain/user/core/repository"
)

var ProviderSet = wire.NewSet(
	// Registry goes here
	NewUserHandlerRegistry,

	// Handlers go here

	// CRUD Repository goes here
	user.NewCoreUserRepository,

	// Subdomains
	account.ProviderSet,
)
