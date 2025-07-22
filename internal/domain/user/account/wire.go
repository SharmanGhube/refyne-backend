package account

import (
	"github.com/google/wire"
	accountRepository "github.com/refynehq/refyne-backend/internal/domain/user/account/repository"
	accountService "github.com/refynehq/refyne-backend/internal/domain/user/account/service"
)

var ProviderSet = wire.NewSet(

	// Handlers go here

	// Services go here
	accountService.NewUserAccountService,

	// Repositories go here
	accountRepository.NewUserSettingsRepository,

// Other dependencies can be added here
)
