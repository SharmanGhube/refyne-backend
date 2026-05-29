package user

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// --- MockCoreUserRepository ---

// MockCoreUserRepository is a hand-written test double for CoreUserRepository.
// Configure return values per-test by setting the exported fields.
type MockCoreUserRepository struct {
	// Return values for GetUserByID
	GetUserByIDUser  *userModels.User
	GetUserByIDError *errors.AppError

	// Return values for GetUserByEmail
	GetUserByEmailUser  *userModels.User
	GetUserByEmailError *errors.AppError

	// Return values for UserExistsByUsername
	UserExistsByUsernameResult bool
	UserExistsByUsernameError  *errors.AppError

	// Return values for UserExistsByEmail
	UserExistsByEmailResult bool
	UserExistsByEmailError  *errors.AppError

	// Return values for UserExistsByID
	UserExistsByIDResult bool
	UserExistsByIDError  *errors.AppError

	// Return values for UpdateUser
	UpdateUserResult *userModels.User
	UpdateUserError  *errors.AppError

	// Return values for simple error-only methods
	CreateUserError          *errors.AppError
	UpdateLastLoginError     *errors.AppError
	VerifyUserError          *errors.AppError
	UpdatePasswordError      *errors.AppError
	SoftDeleteUserError      *errors.AppError
	UpdateOnboardingError    *errors.AppError

	// Call tracking
	GetUserByIDCalls          []string
	UpdateUserCalls           []*userModels.User
	SoftDeleteUserCalls       []string
	UpdateOnboardingCalls     []string
	UserExistsByUsernameCalls []string
}

func (m *MockCoreUserRepository) CreateUser(c *gin.Context, user *userModels.User) *errors.AppError {
	return m.CreateUserError
}

func (m *MockCoreUserRepository) GetUserByID(c *gin.Context, userID string) (*userModels.User, *errors.AppError) {
	m.GetUserByIDCalls = append(m.GetUserByIDCalls, userID)
	return m.GetUserByIDUser, m.GetUserByIDError
}

func (m *MockCoreUserRepository) GetUserByEmail(c *gin.Context, email string) (*userModels.User, *errors.AppError) {
	return m.GetUserByEmailUser, m.GetUserByEmailError
}

func (m *MockCoreUserRepository) UserExistsByID(c *gin.Context, userID string) (bool, *errors.AppError) {
	return m.UserExistsByIDResult, m.UserExistsByIDError
}

func (m *MockCoreUserRepository) UserExistsByEmail(c *gin.Context, email string) (bool, *errors.AppError) {
	return m.UserExistsByEmailResult, m.UserExistsByEmailError
}

func (m *MockCoreUserRepository) UserExistsByUsername(c *gin.Context, username string) (bool, *errors.AppError) {
	m.UserExistsByUsernameCalls = append(m.UserExistsByUsernameCalls, username)
	return m.UserExistsByUsernameResult, m.UserExistsByUsernameError
}

func (m *MockCoreUserRepository) UpdateLastLogin(c *gin.Context, userID string, ipAddress *string, userAgent *string) *errors.AppError {
	return m.UpdateLastLoginError
}

func (m *MockCoreUserRepository) VerifyUser(c *gin.Context, userID string) *errors.AppError {
	return m.VerifyUserError
}

func (m *MockCoreUserRepository) UpdatePassword(c *gin.Context, userID, hashedPassword string) *errors.AppError {
	return m.UpdatePasswordError
}

func (m *MockCoreUserRepository) UpdateUser(c *gin.Context, user *userModels.User) (*userModels.User, *errors.AppError) {
	m.UpdateUserCalls = append(m.UpdateUserCalls, user)
	return m.UpdateUserResult, m.UpdateUserError
}

func (m *MockCoreUserRepository) SoftDeleteUser(c *gin.Context, userID string) *errors.AppError {
	m.SoftDeleteUserCalls = append(m.SoftDeleteUserCalls, userID)
	return m.SoftDeleteUserError
}

func (m *MockCoreUserRepository) UpdateOnboardingStatus(c *gin.Context, userID string, completed bool) *errors.AppError {
	m.UpdateOnboardingCalls = append(m.UpdateOnboardingCalls, userID)
	return m.UpdateOnboardingError
}

func (m *MockCoreUserRepository) GetDB() *sqlx.DB {
	return nil
}

// --- MockSettingsRepository ---

// MockSettingsRepository is a hand-written test double for SettingsRepository.
type MockSettingsRepository struct {
	// Return values for GetSettings
	GetSettingsResult *userModels.UserSettings
	GetSettingsError  *errors.AppError

	// Return values for UpdateSettings
	UpdateSettingsError *errors.AppError

	// Return values for CreateSettings
	CreateSettingsError *errors.AppError

	// Return values for DeleteSettings
	DeleteSettingsError *errors.AppError

	// Call tracking
	GetSettingsCalls    []string
	UpdateSettingsCalls []*userModels.UserSettings
}

func (m *MockSettingsRepository) GetSettings(c *gin.Context, userID string) (*userModels.UserSettings, *errors.AppError) {
	m.GetSettingsCalls = append(m.GetSettingsCalls, userID)
	return m.GetSettingsResult, m.GetSettingsError
}

func (m *MockSettingsRepository) UpdateSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError {
	m.UpdateSettingsCalls = append(m.UpdateSettingsCalls, settings)
	return m.UpdateSettingsError
}

func (m *MockSettingsRepository) CreateSettings(c *gin.Context, settings *userModels.UserSettings) *errors.AppError {
	return m.CreateSettingsError
}

func (m *MockSettingsRepository) DeleteSettings(c *gin.Context, userID string) *errors.AppError {
	return m.DeleteSettingsError
}
