package device

import (
	"github.com/google/wire"
)

// ProviderSet provides device session dependencies
var ProviderSet = wire.NewSet(
	NewDeviceSessionService,
)
