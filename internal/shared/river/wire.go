package riverqueue

import (
	"unsafe"

	"github.com/google/wire"
	emailJobs "github.com/refynehq/refyne-backend/internal/domains/email/jobs"
	instagramJobs "github.com/refynehq/refyne-backend/internal/domains/instagram/jobs"
	"github.com/riverqueue/river"
)

func NewWorkerDependancies(
	emailWorker *emailJobs.EmailWorker,
	webhookWorker *instagramJobs.InstagramWebhookWorker,
	syncMediaWorker *instagramJobs.SyncMediaWorker,
	fetchInsightsWorker *instagramJobs.FetchInsightsWorker,
	refreshTokenWorker *instagramJobs.RefreshTokenWorker,
	processAIWorker *instagramJobs.ProcessAIWorker,
	dailySyncWorker *instagramJobs.DailySyncWorker,
) *WorkerDependancies {
	return &WorkerDependancies{
		EmailWorker:            emailWorker,
		InstagramWebhookWorker: webhookWorker,
		SyncMediaWorker:        syncMediaWorker,
		FetchInsightsWorker:    fetchInsightsWorker,
		RefreshTokenWorker:     refreshTokenWorker,
		ProcessAIWorker:        processAIWorker,
		DailySyncWorker:        dailySyncWorker,
	}
}

// ProvideRiverClientAny provides a type-erased River client for general use
// Uses unsafe pointer conversion as River doesn't support type variance
func ProvideRiverClientAny(service *RiverService) *river.Client[any] {
	// Convert using unsafe - this is safe for River's API as job insertion
	// doesn't actually use the type parameter in a way that matters
	return (*river.Client[any])(unsafe.Pointer(service.GetClient()))
}

var ProviderSet = wire.NewSet(
	NewRiverService,
	NewWorkerDependancies,
	ProvideRiverClientAny,
)
