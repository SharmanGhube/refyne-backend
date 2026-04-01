package user

import (
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	userServices "github.com/refynehq/refyne-backend/internal/domains/user/services"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// UserHandler defines HTTP handlers for user endpoints
type UserHandler interface {
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	GetSettings(c *gin.Context)
	UpdateSettings(c *gin.Context)
	CompleteOnboarding(c *gin.Context)
	DeleteAccount(c *gin.Context)
}

// UserHandlerImpl implements UserHandler
type UserHandlerImpl struct {
	name        string
	logger      *zap.Logger
	userService userServices.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService userServices.UserService) UserHandler {
	return &UserHandlerImpl{
		name:        "UserHandler",
		logger:      logging.GetHandlerLogger("UserHandler"),
		userService: userService,
	}
}

// GetProfile retrieves the current user's profile
// GET /api/user/profile
func (h *UserHandlerImpl) GetProfile(c *gin.Context) {
	requestID := middlewares.GetRequestID(c)
	h.logger.Debug("GetProfile request", zap.String("requestID", requestID))

	userID, exists := middlewares.GetUserID(c)
	if !exists {
		h.logger.Warn("UserID not found in context", zap.String("requestID", requestID))
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	profile, appErr := h.userService.GetUserProfile(c, userID)
	if appErr != nil {
		h.logger.Error("Failed to get user profile", zap.String("requestID", requestID), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}

	h.logger.Info("User profile retrieved", zap.String("requestID", requestID), zap.String("userID", userID))
	c.JSON(200, gin.H{
		"message": "Profile retrieved successfully",
		"data":    profile,
	})
}

// UpdateProfile updates the current user's profile
// PUT /api/user/profile
func (h *UserHandlerImpl) UpdateProfile(c *gin.Context) {
	requestID := middlewares.GetRequestID(c)
	h.logger.Debug("UpdateProfile request", zap.String("requestID", requestID))

	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var updateReq userServices.ProfileUpdateRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		h.logger.Warn("Invalid profile update request", zap.String("requestID", requestID), zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	profile, appErr := h.userService.UpdateUserProfile(c, userID, &updateReq)
	if appErr != nil {
		h.logger.Error("Failed to update user profile", zap.String("requestID", requestID), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}

	h.logger.Info("User profile updated", zap.String("requestID", requestID), zap.String("userID", userID))
	c.JSON(200, gin.H{
		"message": "Profile updated successfully",
		"data":    profile,
	})
}

// GetSettings retrieves the current user's settings
// GET /api/user/settings
func (h *UserHandlerImpl) GetSettings(c *gin.Context) {
	requestID := middlewares.GetRequestID(c)
	h.logger.Debug("GetSettings request", zap.String("requestID", requestID))

	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	settings, appErr := h.userService.GetUserSettings(c, userID)
	if appErr != nil {
		h.logger.Error("Failed to get user settings", zap.String("requestID", requestID), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}

	h.logger.Info("User settings retrieved", zap.String("requestID", requestID), zap.String("userID", userID))
	c.JSON(200, gin.H{
		"message": "Settings retrieved successfully",
		"data":    settings,
	})
}

// UpdateSettings updates the current user's settings
// PUT /api/user/settings
func (h *UserHandlerImpl) UpdateSettings(c *gin.Context) {
	requestID := middlewares.GetRequestID(c)
	h.logger.Debug("UpdateSettings request", zap.String("requestID", requestID))

	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var settings userModels.UserSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		h.logger.Warn("Invalid settings update request", zap.String("requestID", requestID), zap.Error(err))
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	updated, appErr := h.userService.UpdateUserSettings(c, userID, &settings)
	if appErr != nil {
		h.logger.Error("Failed to update user settings", zap.String("requestID", requestID), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}

	h.logger.Info("User settings updated", zap.String("requestID", requestID), zap.String("userID", userID))
	c.JSON(200, gin.H{
		"message": "Settings updated successfully",
		"data":    updated,
	})
}

// CompleteOnboarding marks user's onboarding as completed
// POST /api/user/onboarding
func (h *UserHandlerImpl) CompleteOnboarding(c *gin.Context) {
	requestID := middlewares.GetRequestID(c)
	h.logger.Debug("CompleteOnboarding request", zap.String("requestID", requestID))

	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	appErr := h.userService.CompleteOnboarding(c, userID)
	if appErr != nil {
		h.logger.Error("Failed to complete onboarding", zap.String("requestID", requestID), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}

	h.logger.Info("Onboarding completed", zap.String("requestID", requestID), zap.String("userID", userID))
	c.JSON(200, gin.H{
		"message": "Onboarding completed successfully",
	})
}

// DeleteAccount deletes the current user's account (soft delete)
// DELETE /api/user/account
func (h *UserHandlerImpl) DeleteAccount(c *gin.Context) {
	requestID := middlewares.GetRequestID(c)
	h.logger.Debug("DeleteAccount request", zap.String("requestID", requestID))

	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	appErr := h.userService.DeleteUserAccount(c, userID)
	if appErr != nil {
		h.logger.Error("Failed to delete user account", zap.String("requestID", requestID), zap.Error(appErr))
		c.JSON(appErr.HTTPStatus, appErr)
		return
	}

	h.logger.Info("User account deleted", zap.String("requestID", requestID), zap.String("userID", userID))
	c.JSON(200, gin.H{
		"message": "Account deleted successfully",
	})
}
