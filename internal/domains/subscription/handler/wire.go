package handler

import "github.com/google/wire"

// ProviderSet is the Wire provider set for subscription handlers
var ProviderSet = wire.NewSet(
	NewSubscriptionHandler,
)
