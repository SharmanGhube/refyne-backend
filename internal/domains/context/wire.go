package context

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	// Registry
	NewContextRegistry,

	// Handlers

	// Services

)
