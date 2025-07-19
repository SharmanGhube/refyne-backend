package registry

import (
	auth "github.com/refynehq/refyne-backend/internal/domain/auth/handler"
)

// HandlerRegistry is a struct that holds all the handlers for the application.
// I dont even know why I do this to myself
type HandlerRegistry struct {
	// Auth
	AuthHandler auth.AuthHandler
}

func NewHandlerRegistry(authHandler auth.AuthHandler) *HandlerRegistry {
	return &HandlerRegistry{
		AuthHandler: authHandler,
	}
}
