package jobs

import "github.com/google/wire"

// ProviderSet provides Instagram job workers for River queue
var ProviderSet = wire.NewSet(
	NewInstagramWebhookWorker,
	NewSyncMediaWorker,
	NewFetchInsightsWorker,
	NewRefreshTokenWorker,
	NewProcessAIWorker,
)
