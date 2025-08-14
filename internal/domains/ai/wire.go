package ai

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	// Registry
	NewAIRegistry,

	// Handlers

	// Services

)
