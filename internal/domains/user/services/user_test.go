package user

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestContext creates a minimal gin context for unit tests.
func newTestContext() *gin.Context {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c
}

// --- GetUserProfile ---

func TestGetUserProfile_Success(t *testing.T) {
	mockUser := &userModels.User{
		ID:        "user-1",
		Email:     "test@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
	}

	mockRepo := &MockCoreUserRepository{GetUserByIDUser: mockUser}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	user, appErr := svc.GetUserProfile(c, "user-1")

	require.Nil(t, appErr)
	assert.Equal(t, "user-1", user.ID)
	assert.Equal(t, "Jane", user.FirstName)
	assert.Equal(t, []string{"user-1"}, mockRepo.GetUserByIDCalls)
}

func TestGetUserProfile_NotFound(t *testing.T) {
	mockRepo := &MockCoreUserRepository{GetUserByIDUser: nil}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	user, appErr := svc.GetUserProfile(c, "nonexistent")

	assert.Nil(t, user)
	require.NotNil(t, appErr)
	assert.Equal(t, "USER_NOT_FOUND", appErr.Code)
	assert.Equal(t, errors.ErrorTypeNotFound, appErr.Type)
}

func TestGetUserProfile_RepoError(t *testing.T) {
	repoErr := &errors.AppError{
		Code:    "DB_ERROR",
		Message: "database connection failed",
		Type:    errors.ErrorTypeInternal,
	}

	mockRepo := &MockCoreUserRepository{
		GetUserByIDUser:  nil,
		GetUserByIDError: repoErr,
	}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	user, appErr := svc.GetUserProfile(c, "user-1")

	assert.Nil(t, user)
	require.NotNil(t, appErr)
	assert.Equal(t, "DB_ERROR", appErr.Code)
}

// --- UpdateUserProfile ---

func TestUpdateUserProfile_Success(t *testing.T) {
	existingUser := &userModels.User{
		ID:        "user-1",
		FirstName: "Jane",
		LastName:  "Doe",
		Username:  "janedoe",
	}
	updatedUser := &userModels.User{
		ID:        "user-1",
		FirstName: "Janet",
		LastName:  "Doe",
		Username:  "janedoe",
	}

	mockRepo := &MockCoreUserRepository{
		GetUserByIDUser: existingUser,
		UpdateUserResult: updatedUser,
	}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	result, appErr := svc.UpdateUserProfile(c, "user-1", &ProfileUpdateRequest{
		FirstName: "Janet",
	})

	require.Nil(t, appErr)
	assert.Equal(t, "Janet", result.FirstName)
	assert.Len(t, mockRepo.UpdateUserCalls, 1)
}

func TestUpdateUserProfile_UsernameTaken(t *testing.T) {
	existingUser := &userModels.User{
		ID:       "user-1",
		Username: "janedoe",
	}

	mockRepo := &MockCoreUserRepository{
		GetUserByIDUser:            existingUser,
		UserExistsByUsernameResult: true,
	}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	result, appErr := svc.UpdateUserProfile(c, "user-1", &ProfileUpdateRequest{
		Username: "takenuser",
	})

	assert.Nil(t, result)
	require.NotNil(t, appErr)
	assert.Equal(t, "USERNAME_TAKEN", appErr.Code)
	assert.Equal(t, errors.ErrorTypeValidation, appErr.Type)
}

func TestUpdateUserProfile_SameUsername_NoConflict(t *testing.T) {
	existingUser := &userModels.User{
		ID:       "user-1",
		Username: "janedoe",
	}
	updatedUser := &userModels.User{
		ID:       "user-1",
		Username: "janedoe",
	}

	mockRepo := &MockCoreUserRepository{
		GetUserByIDUser:            existingUser,
		UserExistsByUsernameResult: true, // exists because it's the same user
		UpdateUserResult:           updatedUser,
	}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	result, appErr := svc.UpdateUserProfile(c, "user-1", &ProfileUpdateRequest{
		Username: "janedoe", // same username
	})

	require.Nil(t, appErr)
	assert.Equal(t, "janedoe", result.Username)
}

func TestUpdateUserProfile_UserNotFound(t *testing.T) {
	mockRepo := &MockCoreUserRepository{GetUserByIDUser: nil}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	result, appErr := svc.UpdateUserProfile(c, "nonexistent", &ProfileUpdateRequest{
		FirstName: "Test",
	})

	assert.Nil(t, result)
	require.NotNil(t, appErr)
	assert.Equal(t, "USER_NOT_FOUND", appErr.Code)
}

// --- GetUserSettings ---

func TestGetUserSettings_Success(t *testing.T) {
	expected := &userModels.UserSettings{
		UserID:             "user-1",
		Language:           "en",
		TimeZone:           "America/New_York",
		EmailNotifications: true,
	}

	mockRepo := &MockCoreUserRepository{}
	mockSettings := &MockSettingsRepository{GetSettingsResult: expected}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	settings, appErr := svc.GetUserSettings(c, "user-1")

	require.Nil(t, appErr)
	assert.Equal(t, "America/New_York", settings.TimeZone)
	assert.Equal(t, []string{"user-1"}, mockSettings.GetSettingsCalls)
}

func TestGetUserSettings_NilReturnsDefaults(t *testing.T) {
	mockRepo := &MockCoreUserRepository{}
	mockSettings := &MockSettingsRepository{GetSettingsResult: nil}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	settings, appErr := svc.GetUserSettings(c, "user-1")

	require.Nil(t, appErr)
	require.NotNil(t, settings)
	assert.Equal(t, "user-1", settings.UserID)
	assert.Equal(t, "en", settings.Language)
	assert.Equal(t, "UTC", settings.TimeZone)
	assert.True(t, settings.EmailNotifications)
}

func TestGetUserSettings_RepoError(t *testing.T) {
	repoErr := &errors.AppError{
		Code:    "DB_ERROR",
		Message: "database error",
		Type:    errors.ErrorTypeInternal,
	}

	mockRepo := &MockCoreUserRepository{}
	mockSettings := &MockSettingsRepository{GetSettingsError: repoErr}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	settings, appErr := svc.GetUserSettings(c, "user-1")

	assert.Nil(t, settings)
	require.NotNil(t, appErr)
	assert.Equal(t, "DB_ERROR", appErr.Code)
}

// --- UpdateUserSettings ---

func TestUpdateUserSettings_Success(t *testing.T) {
	input := &userModels.UserSettings{
		Language:           "fr",
		TimeZone:           "Europe/Paris",
		EmailNotifications: false,
	}

	mockRepo := &MockCoreUserRepository{}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	result, appErr := svc.UpdateUserSettings(c, "user-1", input)

	require.Nil(t, appErr)
	assert.Equal(t, "user-1", result.UserID) // UserID should be set by service
	assert.Equal(t, "fr", result.Language)
	assert.Len(t, mockSettings.UpdateSettingsCalls, 1)
}

func TestUpdateUserSettings_RepoError(t *testing.T) {
	repoErr := &errors.AppError{
		Code:    "DB_ERROR",
		Message: "database error",
		Type:    errors.ErrorTypeInternal,
	}

	mockRepo := &MockCoreUserRepository{}
	mockSettings := &MockSettingsRepository{UpdateSettingsError: repoErr}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	result, appErr := svc.UpdateUserSettings(c, "user-1", &userModels.UserSettings{})

	assert.Nil(t, result)
	require.NotNil(t, appErr)
	assert.Equal(t, "DB_ERROR", appErr.Code)
}

// --- CompleteOnboarding ---

func TestCompleteOnboarding_Success(t *testing.T) {
	existingUser := &userModels.User{
		ID:                  "user-1",
		OnboardingCompleted: false,
	}

	mockRepo := &MockCoreUserRepository{GetUserByIDUser: existingUser}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	appErr := svc.CompleteOnboarding(c, "user-1")

	assert.Nil(t, appErr)
	assert.Equal(t, []string{"user-1"}, mockRepo.UpdateOnboardingCalls)
}

func TestCompleteOnboarding_UserNotFound(t *testing.T) {
	mockRepo := &MockCoreUserRepository{GetUserByIDUser: nil}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	appErr := svc.CompleteOnboarding(c, "nonexistent")

	require.NotNil(t, appErr)
	assert.Equal(t, "USER_NOT_FOUND", appErr.Code)
}

// --- DeleteUserAccount ---

func TestDeleteUserAccount_Success(t *testing.T) {
	existingUser := &userModels.User{ID: "user-1"}

	mockRepo := &MockCoreUserRepository{GetUserByIDUser: existingUser}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	appErr := svc.DeleteUserAccount(c, "user-1")

	assert.Nil(t, appErr)
	assert.Equal(t, []string{"user-1"}, mockRepo.SoftDeleteUserCalls)
}

func TestDeleteUserAccount_UserNotFound(t *testing.T) {
	mockRepo := &MockCoreUserRepository{GetUserByIDUser: nil}
	mockSettings := &MockSettingsRepository{}
	svc := NewUserService(mockRepo, mockSettings)

	c := newTestContext()
	appErr := svc.DeleteUserAccount(c, "nonexistent")

	require.NotNil(t, appErr)
	assert.Equal(t, "USER_NOT_FOUND", appErr.Code)
}
