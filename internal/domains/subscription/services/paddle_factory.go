package services

import (
	"fmt"

	"github.com/refynehq/refyne-backend/internal/domains/subscription/config"
	"go.uber.org/zap"
)

// NewPaddleService creates the appropriate Paddle service based on payment mode
func NewPaddleService(cfg *config.PaddleConfig) (PaddleService, error) {
	switch cfg.Mode {
	case config.PaymentModeMock:
		cfg.Logger.Info("Initializing mock Paddle service")
		return NewMockPaddleService(cfg), nil

	case config.PaymentModeSandbox:
		cfg.Logger.Info("Initializing Paddle sandbox service")
		return NewPaddleSandboxService(cfg)

	case config.PaymentModeProduction:
		cfg.Logger.Info("Initializing Paddle production service",
			zap.String("mode", "PRODUCTION"),
		)
		return NewPaddleSandboxService(cfg) // Same implementation, different URL

	default:
		return nil, fmt.Errorf("invalid payment mode: %s", cfg.Mode)
	}
}
