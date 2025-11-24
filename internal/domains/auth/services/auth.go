package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	serviceErrors "github.com/refynehq/refyne-backend/internal/domains/auth/services/errors"
	authUtils "github.com/refynehq/refyne-backend/internal/domains/auth/utils"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	userUtils "github.com/refynehq/refyne-backend/internal/domains/user/utils"
	"github.com/refynehq/refyne-backend/internal/shared/utils"
	"github.com/refynehq/refyne-backend/internal/shared/validation"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

func (s *AuthServiceImpl) RegisterUser(c *gin.Context, firstname, lastname, username, email, password string) *errors.AppError {
	s.logger.Info("Registering User", zap.String("requestID", middlewares.GetRequestID(c)))

	// Comprehensive input validation
	var validationErrors []*validation.ValidationError

	// Validate first name
	if err := s.validator.ValidateName(firstname, "first_name"); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Validate last name
	if err := s.validator.ValidateName(lastname, "last_name"); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Validate username
	if err := s.validator.ValidateUsername(username); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Validate email
	if err := s.validator.ValidateEmail(email); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Validate password
	if err := s.validator.ValidatePassword(password); err != nil {
		validationErrors = append(validationErrors, err)
	}

	// Return all validation errors
	if len(validationErrors) > 0 {
		return validation.NewValidationAppError(c, validationErrors)
	}

	// Sanitize inputs
	firstname = s.validator.SanitizeString(firstname)
	lastname = s.validator.SanitizeString(lastname)
	username = strings.TrimSpace(username) // Username doesn't need HTML escaping
	email = strings.ToLower(strings.TrimSpace(email))

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

	// Log audit event
	s.logRegistration(c, userID, email, username)

	// Send verification email (not welcome email - user must verify first)
	if verifyErr := s.SendVerificationEmail(c, userID, email, username); verifyErr != nil {
		s.logger.Error("Failed to send verification email",
			zap.String("userID", userID),
			zap.String("email", email),
			zap.Error(verifyErr),
		)
		// Don't fail registration if email sending fails
	}

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
	tokenPair, tokenErr := authUtils.GenerateTokenPair(c, user.Username, user.ID, user.Email, user.TokenVersion)
	if tokenErr != nil {
		s.logger.Error("Failed to generate token pair", zap.Error(tokenErr))
		return nil, nil, tokenErr
	}

	// Track device session
	deviceInfo := utils.ExtractDeviceInfo(c)
	if _, err := s.deviceService.CreateOrUpdateSession(c, user.ID, deviceInfo); err != nil {
		s.logger.Warn("Failed to track device session", zap.Error(err))
		// Non-critical, continue
	}

	s.logger.Info("User logged in successfully", zap.String("userID", user.ID))

	// Log audit event
	s.logSuccessfulLogin(c, user.ID, email)

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

	// Check if account is locked
	if lockErr := s.checkAccountLockout(c, user.ID); lockErr != nil {
		return "", lockErr
	}

	// Validate password first
	if isValid, err := authUtils.CheckHash(password, user.PasswordHash); err != nil {
		s.logger.Error("Password hash comparison failed", zap.Error(err))
		s.handleFailedLogin(c, user.ID, email, "password")
		return "", serviceErrors.NewInvalidPasswordError(c, "Invalid password")
	} else if !isValid {
		s.logger.Warn("Invalid password attempt", zap.String("email", email))
		s.handleFailedLogin(c, user.ID, email, "password")
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

	// Log audit event
	s.logOTPGenerated(c, user.ID, email)

	// Send OTP via email asynchronously
	if s.emailService != nil && s.riverClient != nil {
		if err := s.emailService.SendOTP(email, otp); err != nil {
			s.logger.Error("Failed to send OTP email",
				zap.String("email", email),
				zap.Error(err),
			)
			// Don't fail the request if email sending fails
		}
	} else {
		s.logger.Warn("Email service not configured, OTP not sent via email")
	}

	// Return OTP only in development mode for testing
	// In production, OTP is only sent via email
	return "", nil
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

	// Get user from DB first to check lockout
	user, appErr := s.coreUserRepo.GetUserByEmail(c, email)
	if appErr != nil {
		s.logger.Error("Failed to get user by email", zap.Error(appErr))
		return nil, nil, appErr.WithOperation("AuthServiceImpl.VerifyOTPAndLogin - GetUserByEmail")
	}

	// Check if account is locked
	if lockErr := s.checkAccountLockout(c, user.ID); lockErr != nil {
		return nil, nil, lockErr
	}

	// Now validate OTP
	if appErr := otpManager.ValidateOTP(c, email, otp); appErr != nil {
		s.logger.Warn("Invalid OTP attempt", zap.String("email", email))
		s.handleFailedLogin(c, user.ID, email, "otp")
		return nil, nil, appErr
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
	tokenPair, tokenErr := authUtils.GenerateTokenPair(c, user.Username, user.ID, user.Email, user.TokenVersion)
	if tokenErr != nil {
		s.logger.Error("Failed to generate token pair", zap.Error(tokenErr))
		return nil, nil, tokenErr
	}

	// Invalidate the OTP after successful login
	otpManager.InvalidateOTP(email)

	// Clear failed login attempts on successful login
	s.clearFailedAttempts(c, email)

	s.logger.Info("User logged in successfully via OTP", zap.String("userID", user.ID))

	// Log audit event
	s.logOTPVerified(c, user.ID, email)
	s.logSuccessfulLogin(c, user.ID, email)

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
	tokenPair, appErr := authUtils.GenerateTokenPair(c, username, userID, email, user.TokenVersion)
	if appErr != nil {
		return nil, appErr
	}

	s.logger.Info("Token refreshed successfully", zap.String("userID", userID))
	return tokenPair, nil
}

// checkAccountLockout checks if an account is locked and returns appropriate error
func (s *AuthServiceImpl) checkAccountLockout(c *gin.Context, userID string) *errors.AppError {
	locked, lockedUntil, err := s.securityRepo.IsAccountLocked(c, userID)
	if err != nil {
		s.logger.Error("Failed to check account lockout status", zap.Error(err))
		// Don't block login on error, just log it
		return nil
	}

	if locked {
		s.logger.Warn("Attempt to access locked account",
			zap.String("user_id", userID),
			zap.Time("locked_until", lockedUntil),
		)
		return errors.NewAppError(
			c,
			"ACCOUNT_LOCKED",
			fmt.Sprintf("Account is temporarily locked. Please try again after %s", lockedUntil.Format(time.RFC822)),
			errors.ErrorTypeValidation,
			errors.SeverityMedium,
			"auth",
		)
	}

	return nil
}

// handleFailedLogin records failed attempt and locks account if threshold exceeded
func (s *AuthServiceImpl) handleFailedLogin(c *gin.Context, userID, email, attemptType string) {
	ipAddress := c.ClientIP()

	// Record failed attempt
	if err := s.securityRepo.RecordFailedAttempt(c, &userID, email, ipAddress, attemptType); err != nil {
		s.logger.Error("Failed to record failed attempt", zap.Error(err))
	}

	// Check if we should lock the account
	since := time.Now().Add(-15 * time.Minute) // Look at attempts in last 15 minutes
	count, err := s.securityRepo.GetFailedAttemptsCount(c, email, since)
	if err != nil {
		s.logger.Error("Failed to get failed attempts count", zap.Error(err))
		return
	}

	// Lock account after 5 failed attempts in 15 minutes
	if count >= 5 {
		lockDuration := 15 * time.Minute
		if err := s.securityRepo.LockAccount(c, userID, "failed_"+attemptType, lockDuration); err != nil {
			s.logger.Error("Failed to lock account", zap.Error(err))
		} else {
			s.logger.Warn("Account locked due to multiple failed attempts",
				zap.String("user_id", userID),
				zap.String("email", email),
				zap.Int("attempts", count),
			)
		}
	}
}

// clearFailedAttempts clears failed attempts after successful login
func (s *AuthServiceImpl) clearFailedAttempts(c *gin.Context, email string) {
	if err := s.securityRepo.ClearFailedAttempts(c, email); err != nil {
		s.logger.Error("Failed to clear failed attempts", zap.Error(err))
	}
}
