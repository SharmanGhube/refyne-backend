package auth

import (
	"github.com/gin-gonic/gin"
	authRepo "github.com/refynehq/refyne-backend/internal/domains/auth/repository"
	auth "github.com/refynehq/refyne-backend/internal/domains/auth/utils"
	user "github.com/refynehq/refyne-backend/internal/domains/user/core/repository"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type AuthService interface {
	RegisterUser(c *gin.Context, firstname, lastname, username, email, password string) *errors.AppError
	LoginUser(c *gin.Context, email, password string) (*userModels.User, *auth.TokenPair, *errors.AppError)
	RequestOTP(c *gin.Context, email, password string) (string, *errors.AppError)
	VerifyOTPAndLogin(c *gin.Context, email, otp string) (*userModels.User, *auth.TokenPair, *errors.AppError)
	RefreshToken(c *gin.Context, refreshToken string) (*auth.TokenPair, *errors.AppError)
	VerifyAccount(c *gin.Context, token string) *errors.AppError
	Logout(c *gin.Context, token string) *errors.AppError
	LogoutAllDevices(c *gin.Context, userID string) *errors.AppError

	// Password Reset
	RequestPasswordReset(c *gin.Context, email string) *errors.AppError
	ValidateResetToken(c *gin.Context, token string) (*string, *errors.AppError)
	ResetPassword(c *gin.Context, token, newPassword string) *errors.AppError
}

type AuthServiceImpl struct {
	name   string
	logger *zap.Logger

	// Repository dependencies
	coreUserRepo      user.CoreUserRepository
	passwordResetRepo authRepo.PasswordResetRepository

	// Service dependencies

}

func NewAuthService(coreUserRepo user.CoreUserRepository, passwordResetRepo authRepo.PasswordResetRepository) AuthService {
	return &AuthServiceImpl{
		name:              "AuthService",
		logger:            logging.GetServiceLogger("AuthService"),
		coreUserRepo:      coreUserRepo,
		passwordResetRepo: passwordResetRepo,
	}
}
