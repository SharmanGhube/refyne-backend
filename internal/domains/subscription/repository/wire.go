package repository

import "github.com/google/wire"

// ProviderSet is the Wire provider set for subscription repository
var ProviderSet = wire.NewSet(
	NewSubscriptionRepository,
)
