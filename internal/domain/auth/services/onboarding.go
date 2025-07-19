package auth

import (
	"github.com/google/uuid"
	auth "github.com/refynehq/refyne-backend/internal/domain/auth/utils"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

// TODO: Implement these
func (s *authService) RegisterUser(username, password, email string) *errors.AppError {
	s.logger.Info("Registering User", zap.String("username", username))

	// Create User Object
	userID := uuid.New().String()
	hashedPassword, err := auth.GenerateHash(password, 12)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		// Add proper return
	}
	_ = &userModels.User{
		ID:           userID,
		Email:        email,
		Username:     username,
		PasswordHash: hashedPassword, // This should be set after hashing the password
		Status:       "Pending",      // Initial status
		IsActive:     false,          // Initially inactive
		IsVerified:   false,          // Initially not verified
		TimeZone:     "UTC",          // Default timezone
		LastLoginAt:  "",             // No last login yet
		LastLoginIP:  "",             // No last login IP yet
		CreatedAt:    "",             // Set this to current time in actual implementation
		UpdatedAt:    "",             // Set this to current time in actual implementation
		DeletedAt:    "",             // Nullable, for soft deletes
	}

	// Save User to the database

	// Sending a welcome email (riverqueue)

	return nil
}

// TODO: Implement this as well
func (s *authService) LoginUser(username, password string) (string, *errors.AppError) {
	// Here you would typically check the user's credentials against the database.
	// For simplicity, we are just logging the login attempt.
	s.logger.Info("Logging in user", zap.String("username", username))
	return "some-jwt-token", nil
}
