package email

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/domains/email/jobs"
	"github.com/refynehq/refyne-backend/internal/domains/email/service"
)

var ProviderSet = wire.NewSet(
	// Registry
	NewEmailRegistry,

	// Services
	service.NewResendService,
	service.NewEmailService,

	// Workers
	jobs.NewEmailWorker,
)
