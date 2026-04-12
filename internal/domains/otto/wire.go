package otto

import "github.com/google/wire"
import (
	handlers "github.com/refynehq/refyne-backend/internal/domains/otto/handlers"
	services "github.com/refynehq/refyne-backend/internal/domains/otto/services"
)

var ProviderSet = wire.NewSet(
	// Registry
	NewOttoRegistry,

	// Handlers
	handlers.NewOttoHandler,

	// Services
	services.NewConversationService,
)

