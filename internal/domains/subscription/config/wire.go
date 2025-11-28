package config

import "github.com/google/wire"

// ProviderSet is the Wire provider set for subscription configuration
var ProviderSet = wire.NewSet(
	NewPaddleConfig,
)
