package auth

import (
	"github.com/gin-gonic/gin"
	authServices "github.com/refynehq/refyne-backend/internal/domains/auth/services"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type AuthHandler interface {
	Register(c *gin.Context)
	RequestOTP(c *gin.Context)
	VerifyOTP(c *gin.Context)
	LoginWithPassword(c *gin.Context)
	RefreshToken(c *gin.Context)
	Logout(c *gin.Context)
	LogoutAllDevices(c *gin.Context)

	// Email Verification
	VerifyAccount(c *gin.Context)
	ResendVerification(c *gin.Context)

	// Password Reset
	ForgotPassword(c *gin.Context)
	ResetPassword(c *gin.Context)
	ValidateResetToken(c *gin.Context)
}

type AuthHandlerImpl struct {
	name   string
	logger *zap.Logger

	// Dependencies
	authService authServices.AuthService
}

func NewAuthHandler(authService authServices.AuthService) AuthHandler {
	return &AuthHandlerImpl{
		name:        "AuthHandler",
		logger:      logging.GetHandlerLogger("AuthHandler"),
		authService: authService,
	}
}
