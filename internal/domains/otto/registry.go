package otto

import "github.com/refynehq/refyne-backend/internal/domains/otto/handlers"

// OttoRegistry provides access to Otto domain handlers
type OttoRegistry struct {
	Handler *handlers.OttoHandler
}

// NewOttoRegistry creates a new Otto registry
func NewOttoRegistry(handler *handlers.OttoHandler) *OttoRegistry {
	return &OttoRegistry{
		Handler: handler,
	}
}
