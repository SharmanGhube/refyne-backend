package auth

import (
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type AuthHandler interface {
}

type authHandler struct {
	name   string
	logger *zap.Logger
}

func NewAuthHandler() AuthHandler {
	return &authHandler{
		name:   "AuthHandler",
		logger: logging.GetHandlerLogger("AuthHandler"),
	}
}
