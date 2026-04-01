package user

import (
	userHandler "github.com/refynehq/refyne-backend/internal/domains/user/handler"
)

type UserRegistry struct {
	userHandler.UserHandler
}

func NewUserRegistry(userHandler userHandler.UserHandler) *UserRegistry {
	return &UserRegistry{
		UserHandler: userHandler,
	}
}
