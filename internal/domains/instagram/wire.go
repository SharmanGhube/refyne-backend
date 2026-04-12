package instagram

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/config"
	handlers "github.com/refynehq/refyne-backend/internal/domains/instagram/handlers"
	jobs "github.com/refynehq/refyne-backend/internal/domains/instagram/jobs"
	repo "github.com/refynehq/refyne-backend/internal/domains/instagram/repository"
	services "github.com/refynehq/refyne-backend/internal/domains/instagram/services"
)

var ProviderSet = wire.NewSet(
	// Registry
	NewInstagramRegistry,

	// Handlers
	handlers.NewInstagramHandler,

	// Services
	services.NewInstagramOAuthService,
	services.NewInstagramWebhookService,
	services.NewInstagramMediaService,
	services.NewInstagramInsightsService,
	services.NewGeminiService,
	services.NewWebhookDeduplicator,
	services.NewRateLimitChecker,

	// Repository
	repo.NewInstagramAccountRepository,
	repo.NewInstagramMediaRepository,
	repo.NewInstagramInsightsRepository,
	repo.NewInstagramAIRepository,

	// Configuration
	config.NewInstagramConfig,
	config.NewGeminiConfig,

	// Jobs
	jobs.ProviderSet,
)
