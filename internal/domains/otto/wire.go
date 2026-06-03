package otto

import (
	"github.com/google/wire"

	handlers "github.com/refynehq/refyne-backend/internal/domains/otto/handlers"
	repository "github.com/refynehq/refyne-backend/internal/domains/otto/repository"

	services "github.com/refynehq/refyne-backend/internal/domains/otto/services"
)

var ProviderSet = wire.NewSet(
	// Repository
	repository.NewOttoConversationRepository,
	repository.NewOttoMessageRepository,
	repository.NewOttoRepository,

	// Registry
	NewOttoRegistry,

	// Handlers
	handlers.NewOttoHandler,
	handlers.NewOttoSettingsHandler,

	// Services
	services.NewConversationService,
	services.NewOttoAssistantService,
	services.NewOttoService,
)
