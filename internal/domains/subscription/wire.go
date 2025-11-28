package subscription

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/config"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/handler"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/repository"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/services"
)

// ProviderSet is the Wire provider set for the entire subscription domain
var ProviderSet = wire.NewSet(
	// Configuration
	config.ProviderSet,

	// Repository
	repository.ProviderSet,

	// Services
	services.ProviderSet,

	// Handlers
	handler.ProviderSet,

	// Registry
	NewSubscriptionRegistry,
)
