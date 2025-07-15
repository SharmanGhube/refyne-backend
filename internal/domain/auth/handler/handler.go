package auth

import (
	"github.com/gin-gonic/gin"
	authServices "github.com/refynehq/refyne-backend/internal/domain/auth/services"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type authHandler struct {
	name   string
	logger *zap.Logger

	// Dependencies
	authService authServices.AuthService
}

func NewAuthHandler(authService authServices.AuthService) AuthHandler {
	return &authHandler{
		name:        "AuthHandler",
		logger:      logging.GetHandlerLogger("AuthHandler"),
		authService: authService,
	}
}
