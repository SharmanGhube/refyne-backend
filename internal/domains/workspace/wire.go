package workspace

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	// Registry
	NewWorkspaceRegistry,

	// Handlers

	// Services

)
