package subscription

import (
	"github.com/refynehq/refyne-backend/internal/domains/subscription/handler"
)

// SubscriptionRegistry holds all subscription domain handlers
type SubscriptionRegistry struct {
	handler.SubscriptionHandler
}

// NewSubscriptionRegistry creates a new subscription registry
func NewSubscriptionRegistry(subscriptionHandler handler.SubscriptionHandler) *SubscriptionRegistry {
	return &SubscriptionRegistry{
		SubscriptionHandler: subscriptionHandler,
	}
}
