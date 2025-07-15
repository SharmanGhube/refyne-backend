package registry

import (
	auth "github.com/refynehq/refyne-backend/internal/domain/auth/handler"
)

type HandlerRegistry struct {
	// Auth
	AuthHandler auth.AuthHandler
}

func NewHandlerRegistry(authHandler auth.AuthHandler) *HandlerRegistry {
	return &HandlerRegistry{
		AuthHandler: authHandler,
	}
}
