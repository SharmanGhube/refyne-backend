package auth

import (
	authHandler "github.com/refynehq/refyne-backend/internal/domains/auth/handler"
)

type AuthRegistry struct {
	// Add necessary fields for the AuthRegistry
	authHandler.AuthHandler
}

func NewAuthRegistry(authHandler authHandler.AuthHandler) *AuthRegistry {
	return &AuthRegistry{
		AuthHandler: authHandler,
	}
}
