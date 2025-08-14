package notification

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	// Registry
	NewNotificationRegistry,

	// Handlers

	// Services

)
