package auth

import (
	user "github.com/refynehq/refyne-backend/internal/domain/user/repository"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type AuthService interface {
	RegisterUser(username, password, email string) *errors.AppError
	LoginUser(username, password string) (string, *errors.AppError)
}

type authService struct {
	name   string
	logger *zap.Logger

	// Repository Dependencies
	coreUserRepo user.CoreUserRepository
}

func NewAuthService(coreUserRepo user.CoreUserRepository) AuthService {
	return &authService{
		name:         "AuthService",
		logger:       logging.GetServiceLogger("AuthService"),
		coreUserRepo: coreUserRepo,
	}
}
