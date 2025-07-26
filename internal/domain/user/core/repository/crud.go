package user

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	userErrors "github.com/refynehq/refyne-backend/internal/domain/user/core/repository/errors"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

func (r *CoreUserRepositoryImpl) CreateUser(c *gin.Context, user *userModels.User) *errors.AppError {
	r.logger.Info("Creating New User", zap.String("requestID", middlewares.GetRequestID(c)))

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now.Format(time.RFC3339)
	user.UpdatedAt = now.Format(time.RFC3339)

	// Validate user data
	if !user.HasValidEmail() {
		return userErrors.NewUserValidationFailedError(c, "email", "invalid email format")
	}

	// Execute the insert query
	_, err := r.db.NamedExecContext(c.Request.Context(), insertUserQuery, user)
	if err != nil {
		r.logger.Error("Failed to create user", zap.Error(err))
		return userErrors.NewDatabaseError(c, "CreateUser", err)
	}

	r.logger.Info("User created successfully", zap.String("userID", user.ID))
	return nil
}

func (r *CoreUserRepositoryImpl) GetUserByEmail(c *gin.Context, email string) (*userModels.User, *errors.AppError) {
	r.logger.Info("Getting user by email", zap.String("requestID", middlewares.GetRequestID(c)))

	var user userModels.User
	err := r.db.GetContext(c.Request.Context(), &user, selectUserByEmailQuery, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, userErrors.NewUserNotFoundError(c, email)
		}
		r.logger.Error("Failed to get user by email", zap.Error(err))
		return nil, userErrors.NewDatabaseError(c, "GetUserByEmail", err)
	}

	return &user, nil
}

func (r *CoreUserRepositoryImpl) GetUserByUsername(c *gin.Context, username string) (*userModels.User, *errors.AppError) {
	r.logger.Info("Getting user by username", zap.String("requestID", middlewares.GetRequestID(c)))

	var user userModels.User
	err := r.db.GetContext(c.Request.Context(), &user, selectUserByUsernameQuery, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, userErrors.NewUserNotFoundError(c, username)
		}
		r.logger.Error("Failed to get user by username", zap.Error(err))
		return nil, userErrors.NewDatabaseError(c, "GetUserByUsername", err)
	}

	return &user, nil
}

func (r *CoreUserRepositoryImpl) UserExistsByEmail(c *gin.Context, email string) (bool, *errors.AppError) {
	r.logger.Info("Checking if user exists by email", zap.String("requestID", middlewares.GetRequestID(c)))

	var exists bool
	err := r.db.GetContext(c.Request.Context(), &exists, checkUserExistsByEmailQuery, email)
	if err != nil {
		r.logger.Error("Failed to check user existence by email", zap.Error(err))
		return false, userErrors.NewDatabaseError(c, "UserExistsByEmail", err)
	}

	return exists, nil
}

func (r *CoreUserRepositoryImpl) UserExistsByUsername(c *gin.Context, username string) (bool, *errors.AppError) {
	r.logger.Info("Checking if user exists by username", zap.String("requestID", middlewares.GetRequestID(c)))

	var exists bool
	err := r.db.GetContext(c.Request.Context(), &exists, checkUserExistsByUsernameQuery, username)
	if err != nil {
		r.logger.Error("Failed to check user existence by username", zap.Error(err))
		return false, userErrors.NewDatabaseError(c, "UserExistsByUsername", err)
	}

	return exists, nil
}
