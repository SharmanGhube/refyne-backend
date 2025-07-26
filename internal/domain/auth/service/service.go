package auth

import (
	"github.com/gin-gonic/gin"
	account "github.com/refynehq/refyne-backend/internal/domain/user/account/service"
	user "github.com/refynehq/refyne-backend/internal/domain/user/core/repository"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type AuthService interface {
	RegisterUser(c *gin.Context, username, password, email string) *errors.AppError
	LoginUser(c *gin.Context, username, password string) (string, *errors.AppError)
}

type AuthServiceImpl struct {
	name   string
	logger *zap.Logger

	// Dependencies
	CoreUserRepo user.CoreUserRepository

	// Services
	UserAccountService account.UserAccountService
}

func NewAuthService(
	coreUserRepo user.CoreUserRepository,
	userAccountService account.UserAccountService,
) AuthService {
	return &AuthServiceImpl{
		name:               "AuthService",
		logger:             logging.GetServiceLogger("AuthService"),
		CoreUserRepo:       coreUserRepo,
		UserAccountService: userAccountService,
	}
}
