package auth

import (
	"github.com/gin-gonic/gin"
	auth "github.com/refynehq/refyne-backend/internal/domain/auth/service"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
}

type AuthHandlerImpl struct {
	name   string
	logger *zap.Logger

	// Dependencies
	authService auth.AuthService
}

func NewAuthHandler(authService auth.AuthService) AuthHandler {
	return &AuthHandlerImpl{
		name:        "AuthHandler",
		logger:      logging.GetHandlerLogger("authHandler"),
		authService: authService,
	}
}
