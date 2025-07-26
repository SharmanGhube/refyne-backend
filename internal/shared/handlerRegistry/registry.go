package handlerregistry

import (
	"github.com/refynehq/refyne-backend/internal/domain/auth"
	user "github.com/refynehq/refyne-backend/internal/domain/user"
)

type HandlerRegistry struct {
	*auth.AuthRegistry
	*user.UserHandlerRegistry
}

func NewHandlerRegistry(authRegistry *auth.AuthRegistry, userRegistry *user.UserHandlerRegistry) *HandlerRegistry {
	return &HandlerRegistry{
		AuthRegistry:        authRegistry,
		UserHandlerRegistry: userRegistry,
	}
}
