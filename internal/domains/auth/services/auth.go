package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	serviceErrors "github.com/refynehq/refyne-backend/internal/domains/auth/services/errors"
	authUtils "github.com/refynehq/refyne-backend/internal/domains/auth/utils"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	userUtils "github.com/refynehq/refyne-backend/internal/domains/user/utils"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

func (s *AuthServiceImpl) RegisterUser(c *gin.Context, firstname, lastname, username, email, password string) *errors.AppError {
	s.logger.Info("Registering User", zap.String("requestID", middlewares.GetRequestID(c)))

	// Input validation
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if username == "" || email == "" || password == "" {
		return serviceErrors.NewInvalidPasswordError(c, "Username, email, and password are required")
	}

	// Validate password policy
	policy := authUtils.DefaultPasswordPolicy()
	if err := policy.Validate(password); err != nil {
		s.logger.Warn("Password validation failed", zap.Error(err))
		return serviceErrors.NewInvalidPasswordError(c, err.Error())
	}

	// Create User Object for email validation
	tempUser := &userModels.User{Email: email}
	if !tempUser.HasValidEmail() {
		s.logger.Warn("Invalid email format", zap.String("email", email))
		return serviceErrors.NewInvalidEmailFormatError(c, email)
	}

	// Check if user already exists by email
	emailExists, appErr := s.coreUserRepo.UserExistsByEmail(c, email)
	if appErr != nil {
		s.logger.Error("Failed to check if user exists by email", zap.Error(appErr))
		return appErr
	}
	if emailExists {
		s.logger.Warn("User with email already exists", zap.String("email", email))
		return serviceErrors.NewUserAlreadyExistsError(c, "email", email)
	}

	// Check if user already exists by username
	usernameExists, appErr := s.coreUserRepo.UserExistsByUsername(c, username)
	if appErr != nil {
		s.logger.Error("Failed to check if user exists by username", zap.Error(appErr))
		return appErr
	}
	if usernameExists {
		s.logger.Warn("User with username already exists", zap.String("username", username))
		return serviceErrors.NewUserAlreadyExistsError(c, "username", username)
	}

	// Hash password
	hashedPassword, err := authUtils.GenerateHash(password, 12)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return serviceErrors.NewPasswordHashingFailedError(c, err)
	}

	userID := uuid.New().String()
	user := &userModels.User{
		ID:           userID,
		FirstName:    firstname,
		LastName:     lastname,
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		IsVerified:   false,
		Status:       "inactive",
		IsActive:     false,

		DeletedAt: nil,
	}

	// Save User to the database
	if appErr := s.coreUserRepo.CreateUser(c, user); appErr != nil {
		s.logger.Error("Failed to create user", zap.Error(appErr))
		return serviceErrors.NewUserCreationFailedError(c, appErr)
	}

	s.logger.Info("User registered successfully", zap.String("userID", userID), zap.String("username", username))

	// TODO Create Default User Settings

	return nil
}

func (s *AuthServiceImpl) LoginUser(c *gin.Context, email, password string) (*userModels.User, *authUtils.TokenPair, *errors.AppError) {
	s.logger.Info("Logging in User", zap.String("requestID", middlewares.GetRequestID(c)))

	// Validate Input parameters
	if !userUtils.CheckValidEmail(email) {
		return nil, nil, serviceErrors.NewInvalidEmailFormatError(c, email)
	}

	// Check if user exists (by email)
	userExists, appErr := s.coreUserRepo.UserExistsByEmail(c, email)
	if appErr != nil {
		s.logger.Error("Failed to check if user exists by email", zap.Error(appErr))
		return nil, nil, appErr
	}
	if !userExists {
		s.logger.Warn("User with email does not exist", zap.String("email", email))
		return nil, nil, serviceErrors.NewUserNotFoundError(c, email)
	}

	// Get user from DB
	user, appErr := s.coreUserRepo.GetUserByEmail(c, email)
	if appErr != nil {
		s.logger.Error("Failed to get user by email", zap.Error(appErr))
		return nil, nil, appErr.WithOperation("AuthServiceImpl.LoginUser - GetUserByEmail")
	}

	// Compare password hashes
	if isValid, err := authUtils.CheckHash(password, user.PasswordHash); err != nil {
		s.logger.Error("Password hash comparison failed", zap.Error(err))
		return nil, nil, serviceErrors.NewInvalidPasswordError(c, "Invalid password")
	} else if !isValid {
		s.logger.Warn("Invalid password attempt", zap.String("email", email))
		return nil, nil, serviceErrors.NewInvalidPasswordError(c, "Invalid password")
	}

	// Check account status (is_active, is_verified, etc.)
	if !user.IsActive {
		s.logger.Warn("Attempt to login to inactive account", zap.String("email", email))
		return nil, nil, serviceErrors.NewUserNotActiveError(c, email)
	}
	if !user.IsVerified {
		s.logger.Warn("Attempt to login to unverified account", zap.String("email", email))
		return nil, nil, serviceErrors.NewUserNotVerifiedError(c, email)
	}

	// Update last login timestamp and IP
	if appErr := s.coreUserRepo.UpdateLastLogin(c, user.ID, nil, nil); appErr != nil {
		s.logger.Error("Failed to update last login info", zap.Error(appErr))
		// Not a critical error, so we don't return
	}

	// Generate JWT token
	tokenPair, tokenErr := authUtils.GenerateTokenPair(c, user.Username, user.ID, user.Email)
	if tokenErr != nil {
		s.logger.Error("Failed to generate token pair", zap.Error(tokenErr))
		return nil, nil, tokenErr
	}

	s.logger.Info("User logged in successfully", zap.String("userID", user.ID))

	return user, tokenPair, nil
}

// RequestOTP validates user credentials and generates an OTP
func (s *AuthServiceImpl) RequestOTP(c *gin.Context, email, password string) (string, *errors.AppError) {
	s.logger.Info("Requesting OTP for user", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email))

	// Validate Input parameters
	if !userUtils.CheckValidEmail(email) {
		return "", serviceErrors.NewInvalidEmailFormatError(c, email)
	}

	if password == "" {
		return "", serviceErrors.NewInvalidPasswordError(c, "Password is required")
	}

	// Check if user exists (by email)
	userExists, appErr := s.coreUserRepo.UserExistsByEmail(c, email)
	if appErr != nil {
		s.logger.Error("Failed to check if user exists by email", zap.Error(appErr))
		return "", appErr
	}
	if !userExists {
		s.logger.Warn("User with email does not exist", zap.String("email", email))
		return "", serviceErrors.NewUserNotFoundError(c, email)
	}

	// Get user from DB
	user, appErr := s.coreUserRepo.GetUserByEmail(c, email)
	if appErr != nil {
		s.logger.Error("Failed to get user by email", zap.Error(appErr))
		return "", appErr.WithOperation("AuthServiceImpl.RequestOTP - GetUserByEmail")
	}

	// Validate password first
	if isValid, err := authUtils.CheckHash(password, user.PasswordHash); err != nil {
		s.logger.Error("Password hash comparison failed", zap.Error(err))
		return "", serviceErrors.NewInvalidPasswordError(c, "Invalid password")
	} else if !isValid {
		s.logger.Warn("Invalid password attempt", zap.String("email", email))
		return "", serviceErrors.NewInvalidPasswordError(c, "Invalid password")
	}

	// Check account status (is_active, is_verified, etc.)
	// Check if user is verified
	if !user.IsVerified {
		s.logger.Warn("Attempt to request OTP for unverified account", zap.String("email", email))
		return "", serviceErrors.NewUserNotVerifiedError(c, email)
	}

	// Check if user is active
	if !user.IsActive {
		s.logger.Warn("Attempt to request OTP for inactive account", zap.String("email", email))
		return "", serviceErrors.NewUserNotActiveError(c, email)
	}

	// Check if user status is active
	if user.Status != "active" {
		s.logger.Warn("Attempt to request OTP for account with non-active status", zap.String("email", email), zap.String("status", user.Status))
		return "", serviceErrors.NewUserNotActiveError(c, email)
	}

	// Generate OTP
	otpManager := authUtils.GetOTPManager()
	otp, err := otpManager.GenerateOTP()
	if err != nil {
		s.logger.Error("Failed to generate OTP", zap.Error(err))
		return "", serviceErrors.NewInternalServerError(c, "Failed to generate OTP")
	}

	// Store OTP in memory (invalidates any existing OTP for this email)
	otpManager.StoreOTP(email, otp)

	s.logger.Info("OTP generated successfully", zap.String("userID", user.ID), zap.String("email", email))

	// TODO: Send OTP via email service
	// For debug purposes, return the OTP
	return otp, nil
}

// VerifyOTPAndLogin validates OTP and logs in the user
func (s *AuthServiceImpl) VerifyOTPAndLogin(c *gin.Context, email, otp string) (*userModels.User, *authUtils.TokenPair, *errors.AppError) {
	s.logger.Info("Verifying OTP and logging in user", zap.String("requestID", middlewares.GetRequestID(c)), zap.String("email", email))

	// Validate Input parameters
	if !userUtils.CheckValidEmail(email) {
		return nil, nil, serviceErrors.NewInvalidEmailFormatError(c, email)
	}

	if otp == "" {
		return nil, nil, serviceErrors.NewInvalidOTPError(c, "OTP is required")
	}

	// Validate OTP
	otpManager := authUtils.GetOTPManager()
	if appErr := otpManager.ValidateOTP(c, email, otp); appErr != nil {
		s.logger.Warn("Invalid OTP attempt", zap.String("email", email))
		return nil, nil, appErr
	}

	// Get user from DB (we know they exist since OTP was generated)
	user, appErr := s.coreUserRepo.GetUserByEmail(c, email)
	if appErr != nil {
		s.logger.Error("Failed to get user by email", zap.Error(appErr))
		return nil, nil, appErr.WithOperation("AuthServiceImpl.VerifyOTPAndLogin - GetUserByEmail")
	}

	// Double check account status (should be fine since OTP was generated)
	// Check if user is verified
	if !user.IsVerified {
		s.logger.Warn("Attempt to login to unverified account via OTP", zap.String("email", email))
		return nil, nil, serviceErrors.NewUserNotVerifiedError(c, email)
	}

	// Check if user is active
	if !user.IsActive {
		s.logger.Warn("Attempt to login to inactive account via OTP", zap.String("email", email))
		return nil, nil, serviceErrors.NewUserNotActiveError(c, email)
	}

	// Check if user status is active
	if user.Status != "active" {
		s.logger.Warn("Attempt to login to account with non-active status via OTP", zap.String("email", email), zap.String("status", user.Status))
		return nil, nil, serviceErrors.NewUserNotActiveError(c, email)
	}

	// Update last login timestamp and IP
	if appErr := s.coreUserRepo.UpdateLastLogin(c, user.ID, nil, nil); appErr != nil {
		s.logger.Error("Failed to update last login info", zap.Error(appErr))
		// Not a critical error, so we don't return
	}

	// Generate JWT token
	tokenPair, tokenErr := authUtils.GenerateTokenPair(c, user.Username, user.ID, user.Email)
	if tokenErr != nil {
		s.logger.Error("Failed to generate token pair", zap.Error(tokenErr))
		return nil, nil, tokenErr
	}

	// Invalidate the OTP after successful login
	otpManager.InvalidateOTP(email)

	s.logger.Info("User logged in successfully via OTP", zap.String("userID", user.ID))

	return user, tokenPair, nil
}

// RefreshToken validates refresh token and generates new token pair
func (s *AuthServiceImpl) RefreshToken(c *gin.Context, refreshToken string) (*authUtils.TokenPair, *errors.AppError) {
	s.logger.Info("Refreshing token", zap.String("requestID", middlewares.GetRequestID(c)))

	// Validate refresh token
	claims, err := authUtils.ValidateAndExtractToken(refreshToken)
	if err != nil {
		s.logger.Warn("Invalid refresh token", zap.Error(err))
		return nil, serviceErrors.NewInvalidTokenError(c, "Invalid refresh token")
	}

	// Extract user information from claims
	userID := claims.UserID
	username := claims.Username
	email := claims.Email

	if userID == "" || username == "" || email == "" {
		return nil, serviceErrors.NewInvalidTokenError(c, "Invalid token claims")
	}

	// Verify user still exists and is active
	user, appErr := s.coreUserRepo.GetUserByEmail(c, email)
	if appErr != nil {
		return nil, appErr
	}

	// Verify the user ID matches (additional security check)
	if user.ID != userID {
		s.logger.Warn("User ID mismatch in token",
			zap.String("tokenUserID", userID),
			zap.String("dbUserID", user.ID))
		return nil, serviceErrors.NewInvalidTokenError(c, "Invalid token")
	}

	// Check user status - same comprehensive checks as in login
	if !user.IsVerified {
		return nil, serviceErrors.NewUserNotVerifiedError(c, user.Email)
	}

	if !user.IsActive {
		return nil, serviceErrors.NewUserNotActiveError(c, user.Email)
	}

	if user.Status != "active" {
		return nil, serviceErrors.NewUserNotActiveError(c, user.Email)
	}

	// Generate new token pair
	tokenPair, appErr := authUtils.GenerateTokenPair(c, username, userID, email)
	if appErr != nil {
		return nil, appErr
	}

	s.logger.Info("Token refreshed successfully", zap.String("userID", userID))
	return tokenPair, nil
}

// VerifyAccount verifies user account using JWT token
func (s *AuthServiceImpl) VerifyAccount(c *gin.Context, token string) *errors.AppError {
	s.logger.Info("Verifying user account", zap.String("requestID", middlewares.GetRequestID(c)))

	// Validate token
	claims, err := authUtils.ValidateAndExtractToken(token)
	if err != nil {
		s.logger.Warn("Invalid verification token", zap.Error(err))
		return serviceErrors.NewInvalidTokenError(c, "Invalid verification token")
	}

	// Extract user information from claims
	userID := claims.UserID
	email := claims.Email

	if userID == "" || email == "" {
		return serviceErrors.NewInvalidTokenError(c, "Invalid token claims")
	}

	// Get user from database
	user, appErr := s.coreUserRepo.GetUserByEmail(c, email)
	if appErr != nil {
		return appErr
	}

	// Verify user ID matches (security check)
	if user.ID != userID {
		s.logger.Warn("User ID mismatch in verification token",
			zap.String("tokenUserID", userID),
			zap.String("dbUserID", user.ID))
		return serviceErrors.NewInvalidTokenError(c, "Invalid verification token")
	}

	// Check if account is already verified
	if user.IsVerified && user.IsActive && user.Status == "active" {
		return serviceErrors.NewAccountAlreadyVerifiedError(c, user.Email)
	}

	// Update user verification status
	appErr = s.coreUserRepo.VerifyUser(c, userID)
	if appErr != nil {
		return appErr
	}

	s.logger.Info("User account verified successfully",
		zap.String("userID", userID),
		zap.String("email", user.Email))

	return nil
}
