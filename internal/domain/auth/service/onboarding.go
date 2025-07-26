package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	authErrors "github.com/refynehq/refyne-backend/internal/domain/auth/service/errors"
	auth "github.com/refynehq/refyne-backend/internal/domain/auth/utils"
	userModels "github.com/refynehq/refyne-backend/internal/domain/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

func (s *AuthServiceImpl) RegisterUser(c *gin.Context, username, password, email string) *errors.AppError {
	s.logger.Info("Registering User", zap.String("username", username), zap.String("email", email))

	// Input validation
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if username == "" || email == "" || password == "" {
		return authErrors.NewInvalidPasswordError(c, "Username, email, and password are required")
	}

	// Validate password policy
	policy := auth.DefaultPasswordPolicy()
	if err := policy.Validate(password); err != nil {
		s.logger.Warn("Password validation failed", zap.Error(err))
		return authErrors.NewInvalidPasswordError(c, err.Error())
	}

	// Create User Object for email validation
	tempUser := &userModels.User{Email: email}
	if !tempUser.HasValidEmail() {
		s.logger.Warn("Invalid email format", zap.String("email", email))
		return authErrors.NewInvalidEmailError(c, email)
	}

	// Check if user already exists by email
	emailExists, appErr := s.CoreUserRepo.UserExistsByEmail(c, email)
	if appErr != nil {
		s.logger.Error("Failed to check if user exists by email", zap.Error(appErr))
		return appErr
	}
	if emailExists {
		s.logger.Warn("User with email already exists", zap.String("email", email))
		return authErrors.NewUserAlreadyExistsError(c, "email", email)
	}

	// Check if user already exists by username
	usernameExists, appErr := s.CoreUserRepo.UserExistsByUsername(c, username)
	if appErr != nil {
		s.logger.Error("Failed to check if user exists by username", zap.Error(appErr))
		return appErr
	}
	if usernameExists {
		s.logger.Warn("User with username already exists", zap.String("username", username))
		return authErrors.NewUserAlreadyExistsError(c, "username", username)
	}

	// Hash password
	hashedPassword, err := auth.GenerateHash(password, 12)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return authErrors.NewPasswordHashingFailedError(c, err)
	}

	// Create User Object
	userID := uuid.New().String()
	user := &userModels.User{
		ID:           userID,
		Email:        email,
		Username:     username,
		PasswordHash: hashedPassword,
		Status:       "Pending",
		IsActive:     false,
		IsVerified:   false,
		LastLoginAt:  "",
		LastLoginIP:  "",
		DeletedAt:    "",
	}

	// Save User to the database
	if appErr := s.CoreUserRepo.CreateUser(c, user); appErr != nil {
		s.logger.Error("Failed to create user", zap.Error(appErr))
		return authErrors.NewUserCreationFailedError(c, appErr)
	}

	s.logger.Info("User registered successfully", zap.String("userID", userID), zap.String("username", username))

	// Create default user settings
	if appErr := s.UserAccountService.CreateDefaultUserSettings(c, userID); appErr != nil {
		s.logger.Error("Failed to create default user settings",
			zap.String("userID", userID),
			zap.Error(appErr))
		// Note: We don't return error here as user creation was successful
		// This is a non-critical failure that can be handled later
		s.logger.Warn("User created but default settings creation failed",
			zap.String("userID", userID))
	} else {
		s.logger.Info("Default user settings created successfully", zap.String("userID", userID))
	}

	// TODO: Sending a welcome email (riverqueue)

	return nil
}

func (s *AuthServiceImpl) LoginUser(c *gin.Context, username, password string) (string, *errors.AppError) {
	// Here you would typically check the user's credentials against the database.
	// For simplicity, we are just logging the login attempt.
	s.logger.Info("Logging in user", zap.String("username", username))
	return "some-jwt-token", nil
}
