package services

import "github.com/google/wire"

// ProviderSet is the Wire provider set for subscription services
var ProviderSet = wire.NewSet(
	NewPaddleService,
	NewWebhookService,
)
