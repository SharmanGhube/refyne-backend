package auth

import (
	"github.com/gin-gonic/gin"
	accountService "github.com/refynehq/refyne-backend/internal/domain/user/account/service"
	user "github.com/refynehq/refyne-backend/internal/domain/user/repository"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type AuthService interface {
	RegisterUser(c *gin.Context, username, password, email string) *errors.AppError
	LoginUser(c *gin.Context, username, password string) (string, *errors.AppError)
}

type authService struct {
	name   string
	logger *zap.Logger

	// Repository Dependencies
	coreUserRepo user.CoreUserRepository

	// Service Dependencies
	userAccountService accountService.UserAccountService
}

func NewAuthService(coreUserRepo user.CoreUserRepository, userAccountService accountService.UserAccountService) AuthService {
	return &authService{
		name:               "AuthService",
		logger:             logging.GetServiceLogger("AuthService"),
		coreUserRepo:       coreUserRepo,
		userAccountService: userAccountService,
	}
}
