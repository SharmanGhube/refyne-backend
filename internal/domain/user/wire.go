package user

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/domain/user/account"
	userCoreRepository "github.com/refynehq/refyne-backend/internal/domain/user/repository"
)

var ProviderSet = wire.NewSet(

	// CRUD repositories go here
	userCoreRepository.NewCoreUserRepository,

	// SubDomains ProviderSets
	account.ProviderSet,

// Other dependencies can be added here
)
