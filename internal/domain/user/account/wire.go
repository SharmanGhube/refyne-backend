package account

import (
	"github.com/google/wire"
	accountHandler "github.com/refynehq/refyne-backend/internal/domain/user/account/handler"
	accountRepository "github.com/refynehq/refyne-backend/internal/domain/user/account/repository"
	accountService "github.com/refynehq/refyne-backend/internal/domain/user/account/service"
)

var ProviderSet = wire.NewSet(

	// Handlers go here
	accountHandler.NewUserAccountHandler,

	// Services go here
	accountService.NewUserAccountService,

	// Repositories go here
	accountRepository.NewUserSettingsRepository,

// Other dependencies can be added here
)
