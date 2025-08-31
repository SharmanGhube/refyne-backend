package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type CoreUserRepository interface {
	// CRUD operations for user
	CreateUser(c *gin.Context, user *userModels.User) *errors.AppError

	GetUserByID(c *gin.Context, userID string) (*userModels.User, *errors.AppError)
	GetUserByEmail(c *gin.Context, email string) (*userModels.User, *errors.AppError)

	UserExistsByID(c *gin.Context, userID string) (bool, *errors.AppError)
	UserExistsByEmail(c *gin.Context, email string) (bool, *errors.AppError)
	UserExistsByUsername(c *gin.Context, username string) (bool, *errors.AppError)

	UpdateLastLogin(c *gin.Context, userID string, ipAddress *string, userAgent *string) *errors.AppError
	VerifyUser(c *gin.Context, userID string) *errors.AppError
}

type CoreUserRepositoryImpl struct {
	name   string
	db     *sqlx.DB
	logger *zap.Logger
}

func NewCoreUserRepository(db *sqlx.DB) CoreUserRepository {
	return &CoreUserRepositoryImpl{
		name:   "CoreUserRepository",
		db:     db,
		logger: logging.GetRepositoryLogger("CoreUserRepository"),
	}
}
