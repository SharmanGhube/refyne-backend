package auth

import (
	"github.com/google/wire"
	"github.com/refynehq/refyne-backend/internal/config"
	authHandler "github.com/refynehq/refyne-backend/internal/domains/auth/handler"
	authRepo "github.com/refynehq/refyne-backend/internal/domains/auth/repository"
	authService "github.com/refynehq/refyne-backend/internal/domains/auth/services"
	"github.com/refynehq/refyne-backend/internal/shared/audit"
	"github.com/refynehq/refyne-backend/internal/shared/device"
	"github.com/refynehq/refyne-backend/internal/shared/validation"
)

// ProvideFrontendURL extracts frontend URL from config for auth service
func ProvideFrontendURL(cfg *config.Config) string {
	return cfg.FrontendURL
}

var ProviderSet = wire.NewSet(
	// Registry
	NewAuthRegistry,

	// Handlers
	authHandler.NewAuthHandler,

	// Services
	authService.NewAuthService,
	ProvideFrontendURL,

	// Repositories
	authRepo.NewPasswordResetRepository,
	authRepo.NewVerificationRepository,
	authRepo.NewAccountSecurityRepository,

	// Audit
	audit.ProvideAuditLogger,

	// Device tracking
	device.ProviderSet,

	// Validation
	validation.ProviderSet,
)
