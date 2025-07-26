package auth

import auth "github.com/refynehq/refyne-backend/internal/domain/auth/handler"

type AuthRegistry struct {
	// Handlers
	AuthHandler auth.AuthHandler
}

func NewAuthRegistry(AuthHandler auth.AuthHandler) *AuthRegistry {
	return &AuthRegistry{
		AuthHandler: AuthHandler,
	}
}
