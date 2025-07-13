package riverqueue

import "github.com/google/wire"

func NewWorkerDependancies() *WorkerDependancies {
	return &WorkerDependancies{
		// Add dependencies needed for the worker here
	}
}

var ProviderSet = wire.NewSet(
	NewRiverService,
	NewWorkerDependancies,
)
