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

	// Registry
	NewOttoRegistry,

	// Handlers
	handlers.NewOttoHandler,

	// Services
	services.NewConversationService,
	services.NewOttoAssistantService,
)
