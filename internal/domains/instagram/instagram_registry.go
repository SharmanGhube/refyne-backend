package instagram

import (
	"github.com/refynehq/refyne-backend/internal/domains/instagram/handlers"
)

// InstagramRegistry manages Instagram domain routes
type InstagramRegistry struct {
	*handlers.InstagramHandler
}

// NewInstagramRegistry creates a new Instagram registry
func NewInstagramRegistry(handler *handlers.InstagramHandler) *InstagramRegistry {
	return &InstagramRegistry{
		InstagramHandler: handler,
	}
}
