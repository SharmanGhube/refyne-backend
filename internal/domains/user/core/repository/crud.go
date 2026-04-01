package user

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	coreRepoErrors "github.com/refynehq/refyne-backend/internal/domains/user/core/repository/errors"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

func (r *CoreUserRepositoryImpl) CreateUser(c *gin.Context, user *userModels.User) *errors.AppError {
	r.logger.Info("Creating new User", zap.String("requestID", middlewares.GetRequestID(c)))

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Execute insert query
	_, err := r.db.NamedExecContext(c.Request.Context(), insertUserQuery, user)
	if err != nil {
		r.logger.Error("Failed to create user", zap.String("requestID", middlewares.GetRequestID(c)), zap.Error(err))
		return coreRepoErrors.NewDatabaseError(c, "CreateUser", err)
	}

	r.logger.Info("User created successfully", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", user.ID))
	return nil
}

func (r *CoreUserRepositoryImpl) GetUserByID(c *gin.Context, userID string) (*userModels.User, *errors.AppError) {
	r.logger.Info("Getting user by ID", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID))

	var user userModels.User
	err := r.db.GetContext(c.Request.Context(), &user, selectUserByIDQuery, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("User not found", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID))
			return nil, nil // Return nil user with no error to indicate user not found
		}
		r.logger.Error("Failed to get user by ID", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID), zap.Error(err))
		return nil, coreRepoErrors.NewDatabaseError(c, "GetUserByID", err)
	}

	r.logger.Info("User retrieved successfully", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID))
	return &user, nil
}

func (r *CoreUserRepositoryImpl) GetUserByEmail(c *gin.Context, email string) (*userModels.User, *errors.AppError) {
	r.logger.Info("Getting user by email", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email))

	var user userModels.User
	err := r.db.GetContext(c.Request.Context(), &user, selectUserByEmailQuery, email)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("User not found by email", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email))
			return nil, nil // Return nil user with no error to indicate user not found
		}
		r.logger.Error("Failed to get user by email", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email), zap.Error(err))
		return nil, coreRepoErrors.NewDatabaseError(c, "GetUserByEmail", err)
	}

	r.logger.Info("User retrieved successfully by email", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email))
	return &user, nil
}

func (r *CoreUserRepositoryImpl) UserExistsByID(c *gin.Context, userID string) (bool, *errors.AppError) {
	r.logger.Info("Checking if user exists by ID", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID))

	var count int
	err := r.db.GetContext(c.Request.Context(), &count, "SELECT COUNT(*) FROM users WHERE id = $1 AND deleted_at IS NULL", userID)
	if err != nil {
		r.logger.Error("Failed to check user existence by ID", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID), zap.Error(err))
		return false, coreRepoErrors.NewDatabaseError(c, "UserExistsByID", err)
	}

	exists := count > 0
	r.logger.Info("User existence check completed", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID), zap.Bool("exists", exists))
	return exists, nil
}

func (r *CoreUserRepositoryImpl) UserExistsByEmail(c *gin.Context, email string) (bool, *errors.AppError) {
	r.logger.Info("Checking if user exists by email", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email))

	var count int
	err := r.db.GetContext(c.Request.Context(), &count, "SELECT COUNT(*) FROM users WHERE email = $1 AND deleted_at IS NULL", email)
	if err != nil {
		r.logger.Error("Failed to check user existence by email", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email), zap.Error(err))
		return false, coreRepoErrors.NewDatabaseError(c, "UserExistsByEmail", err)
	}

	exists := count > 0
	r.logger.Info("User existence check completed", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email), zap.Bool("exists", exists))
	return exists, nil
}

func (r *CoreUserRepositoryImpl) UserExistsByUsername(c *gin.Context, username string) (bool, *errors.AppError) {
	r.logger.Info("Checking if user exists by username", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("username", username))

	var count int
	err := r.db.GetContext(c.Request.Context(), &count, "SELECT COUNT(*) FROM users WHERE username = $1 AND deleted_at IS NULL", username)
	if err != nil {
		r.logger.Error("Failed to check user existence by username", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("username", username), zap.Error(err))
		return false, coreRepoErrors.NewDatabaseError(c, "UserExistsByUsername", err)
	}

	exists := count > 0
	r.logger.Info("User existence check completed", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("username", username), zap.Bool("exists", exists))
	return exists, nil
}

func (r *CoreUserRepositoryImpl) UpdateLastLogin(c *gin.Context, userID string, ipAddress *string, userAgent *string) *errors.AppError {
	r.logger.Info("Updating last login info", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID))

	// Use a simple query with the defined query constant
	_, err := r.db.ExecContext(c.Request.Context(), updateUserLoginInfoQuery, userID, time.Now(), ipAddress)
	if err != nil {
		r.logger.Error("Failed to update last login info", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID), zap.Error(err))
		return coreRepoErrors.NewDatabaseError(c, "UpdateLastLogin", err)
	}

	// Note: userAgent parameter is ignored as the database doesn't have a user_agent field
	_ = userAgent // Suppress unused parameter warning

	r.logger.Info("Last login info updated successfully", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID))
	return nil
}

// VerifyUser updates user verification status
func (r *CoreUserRepositoryImpl) VerifyUser(c *gin.Context, userID string) *errors.AppError {
	r.logger.Info("Verifying user",
		zap.String("userID", userID),
		zap.String("requestID", middlewares.GetRequestID(c)))

	// Execute update query to verify user
	result, err := r.db.ExecContext(c.Request.Context(), verifyUserQuery, userID)
	if err != nil {
		r.logger.Error("Failed to verify user",
			zap.Error(err),
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewDatabaseError(c, "VerifyUser", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected",
			zap.Error(err),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewDatabaseError(c, "VerifyUser", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("No user found for verification",
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewUserNotFoundError(c, "User not found for verification")
	}

	r.logger.Info("User verified successfully",
		zap.String("userID", userID),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return nil
}

// UpdatePassword updates a user's password
func (r *CoreUserRepositoryImpl) UpdatePassword(c *gin.Context, userID, hashedPassword string) *errors.AppError {
	r.logger.Info("Updating user password", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("userID", userID))

	query := `UPDATE users 
	          SET password_hash = $1, 
	              last_password_changed_at = CURRENT_TIMESTAMP,
	              token_version = token_version + 1,
	              updated_at = CURRENT_TIMESTAMP 
	          WHERE id = $2`
	result, err := r.db.ExecContext(c.Request.Context(), query, hashedPassword, userID)
	if err != nil {
		r.logger.Error("Failed to update password",
			zap.String("requestID", middlewares.GetRequestID(c)),
			zap.String("userID", userID),
			zap.Error(err))
		return coreRepoErrors.NewDatabaseError(c, "UpdatePassword", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected",
			zap.Error(err),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewDatabaseError(c, "UpdatePassword", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("No user found for password update",
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewUserNotFoundError(c, "User not found for password update")
	}

	r.logger.Info("Password updated successfully",
		zap.String("userID", userID),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return nil
}

// GetDB returns the database connection for direct access when needed
func (r *CoreUserRepositoryImpl) GetDB() *sqlx.DB {
	return r.db
}

// UpdateUser updates user profile information
func (r *CoreUserRepositoryImpl) UpdateUser(c *gin.Context, user *userModels.User) (*userModels.User, *errors.AppError) {
	r.logger.Info("Updating user",
		zap.String("userID", user.ID),
		zap.String("requestID", middlewares.GetRequestID(c)))

	var updatedUser userModels.User
	err := r.db.GetContext(c.Request.Context(), &updatedUser, updateUserQuery,
		user.ID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Status,
		user.IsActive,
		user.IsVerified,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("User not found for update",
				zap.String("userID", user.ID),
				zap.String("requestID", middlewares.GetRequestID(c)))
			return nil, coreRepoErrors.NewUserNotFoundError(c, "User not found for update")
		}
		r.logger.Error("Failed to update user",
			zap.Error(err),
			zap.String("userID", user.ID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return nil, coreRepoErrors.NewDatabaseError(c, "UpdateUser", err)
	}

	r.logger.Info("User updated successfully",
		zap.String("userID", user.ID),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return &updatedUser, nil
}

// SoftDeleteUser marks a user as deleted
func (r *CoreUserRepositoryImpl) SoftDeleteUser(c *gin.Context, userID string) *errors.AppError {
	r.logger.Info("Soft deleting user",
		zap.String("userID", userID),
		zap.String("requestID", middlewares.GetRequestID(c)))

	result, err := r.db.ExecContext(c.Request.Context(), softDeleteUserQuery, userID)
	if err != nil {
		r.logger.Error("Failed to soft delete user",
			zap.Error(err),
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewDatabaseError(c, "SoftDeleteUser", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected",
			zap.Error(err),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewDatabaseError(c, "SoftDeleteUser", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("No user found for deletion",
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewUserNotFoundError(c, "User not found for deletion")
	}

	r.logger.Info("User soft deleted successfully",
		zap.String("userID", userID),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return nil
}

// UpdateOnboardingStatus updates the user's onboarding completion status
func (r *CoreUserRepositoryImpl) UpdateOnboardingStatus(c *gin.Context, userID string, completed bool) *errors.AppError {
	r.logger.Info("Updating onboarding status",
		zap.String("userID", userID),
		zap.Bool("completed", completed),
		zap.String("requestID", middlewares.GetRequestID(c)))

	query := `UPDATE users SET onboarding_completed = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(c.Request.Context(), query, userID, completed)
	if err != nil {
		r.logger.Error("Failed to update onboarding status",
			zap.Error(err),
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewDatabaseError(c, "UpdateOnboardingStatus", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected",
			zap.Error(err),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewDatabaseError(c, "UpdateOnboardingStatus", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("No user found for onboarding update",
			zap.String("userID", userID),
			zap.String("requestID", middlewares.GetRequestID(c)))
		return coreRepoErrors.NewUserNotFoundError(c, "User not found for onboarding update")
	}

	r.logger.Info("Onboarding status updated successfully",
		zap.String("userID", userID),
		zap.Bool("completed", completed),
		zap.String("requestID", middlewares.GetRequestID(c)))
	return nil
}
