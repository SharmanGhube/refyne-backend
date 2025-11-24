package validation

import "github.com/google/wire"

// ProviderSet provides validation dependencies
var ProviderSet = wire.NewSet(
	NewValidator,
)
