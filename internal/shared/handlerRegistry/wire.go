package handlerregistry

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	// Constructor for HandlerRegistry
	NewHandlerRegistry,
)
