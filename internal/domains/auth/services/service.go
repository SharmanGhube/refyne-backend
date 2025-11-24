package auth

import (
	"github.com/gin-gonic/gin"
	authRepo "github.com/refynehq/refyne-backend/internal/domains/auth/repository"
	auth "github.com/refynehq/refyne-backend/internal/domains/auth/utils"
	emailService "github.com/refynehq/refyne-backend/internal/domains/email/service"
	user "github.com/refynehq/refyne-backend/internal/domains/user/core/repository"
	userModels "github.com/refynehq/refyne-backend/internal/domains/user/models"
	"github.com/refynehq/refyne-backend/internal/shared/audit"
	"github.com/refynehq/refyne-backend/internal/shared/device"
	"github.com/refynehq/refyne-backend/internal/shared/validation"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type AuthService interface {
	RegisterUser(c *gin.Context, firstname, lastname, username, email, password string) *errors.AppError
	LoginUser(c *gin.Context, email, password string) (*userModels.User, *auth.TokenPair, *errors.AppError)
	RequestOTP(c *gin.Context, email, password string) (string, *errors.AppError)
	VerifyOTPAndLogin(c *gin.Context, email, otp string) (*userModels.User, *auth.TokenPair, *errors.AppError)
	RefreshToken(c *gin.Context, refreshToken string) (*auth.TokenPair, *errors.AppError)
	Logout(c *gin.Context, token string) *errors.AppError
	LogoutAllDevices(c *gin.Context, userID string) *errors.AppError

	// Password Reset
	RequestPasswordReset(c *gin.Context, email string) *errors.AppError
	ValidateResetToken(c *gin.Context, token string) (*string, *errors.AppError)
	ResetPassword(c *gin.Context, token, newPassword string) *errors.AppError

	// Account Verification
	SendVerificationEmail(c *gin.Context, userID, email, username string) *errors.AppError
	VerifyAccount(c *gin.Context, token string) *errors.AppError
	ResendVerificationEmail(c *gin.Context, email string) *errors.AppError
}

type AuthServiceImpl struct {
	name   string
	logger *zap.Logger

	// Repository dependencies
	coreUserRepo      user.CoreUserRepository
	passwordResetRepo authRepo.PasswordResetRepository
	verificationRepo  authRepo.VerificationRepository
	securityRepo      authRepo.AccountSecurityRepository

	// Service dependencies
	emailService  emailService.EmailService
	riverClient   *river.Client[any]
	auditLogger   *audit.AuditLogger
	deviceService *device.DeviceSessionService
	validator     *validation.Validator
	frontendURL   string
}

func NewAuthService(
	coreUserRepo user.CoreUserRepository,
	passwordResetRepo authRepo.PasswordResetRepository,
	verificationRepo authRepo.VerificationRepository,
	securityRepo authRepo.AccountSecurityRepository,
	emailService emailService.EmailService,
	riverClient *river.Client[any],
	auditLogger *audit.AuditLogger,
	deviceService *device.DeviceSessionService,
	validator *validation.Validator,
	frontendURL string,
) AuthService {
	return &AuthServiceImpl{
		name:              "AuthService",
		logger:            logging.GetServiceLogger("AuthService"),
		coreUserRepo:      coreUserRepo,
		passwordResetRepo: passwordResetRepo,
		verificationRepo:  verificationRepo,
		securityRepo:      securityRepo,
		emailService:      emailService,
		riverClient:       riverClient,
		auditLogger:       auditLogger,
		deviceService:     deviceService,
		validator:         validator,
		frontendURL:       frontendURL,
	}
}
