package user

import account "github.com/refynehq/refyne-backend/internal/domain/user/account/handler"

type UserHandlerRegistry struct {
	AccountHandler account.UserAccountHandler
}

func NewUserHandlerRegistry(accountHandler account.UserAccountHandler) *UserHandlerRegistry {
	return &UserHandlerRegistry{
		AccountHandler: accountHandler,
	}
}
