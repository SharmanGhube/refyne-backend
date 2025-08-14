package otto

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	// Registry
	NewOttoRegistry,

	// Handlers

	// Services

)
