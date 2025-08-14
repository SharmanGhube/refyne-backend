package email

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	// Registry
	NewEmailRegistry,

	// Handlers

	// Services

)
