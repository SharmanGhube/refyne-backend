package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// This repository will have Core CRUD operations for the user domain.
// I love torturing myself, gives me the high without the need for drugs.

type CoreUserRepository interface {
	CreateUser(c *gin.Context, user *userModels.User) *errors.AppError
	GetUserByEmail(c *gin.Context, email string) (*userModels.User, *errors.AppError)
	GetUserByUsername(c *gin.Context, username string) (*userModels.User, *errors.AppError)
	UserExistsByEmail(c *gin.Context, email string) (bool, *errors.AppError)
	UserExistsByUsername(c *gin.Context, username string) (bool, *errors.AppError)
}

type coreUserRepository struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

func NewCoreUserRepository(db *sqlx.DB) CoreUserRepository {
	return &coreUserRepository{
		name:   "CoreUserRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("CoreUserRepository"),
	}
}
