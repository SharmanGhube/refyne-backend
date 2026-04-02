package config

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
)

// PaymentMode represents the payment processing mode
type PaymentMode string

const (
	PaymentModeMock       PaymentMode = "mock"
	PaymentModeSandbox    PaymentMode = "sandbox"
	PaymentModeProduction PaymentMode = "production"
)

// PaddleConfig holds all Paddle-related configuration
type PaddleConfig struct {
	// Payment mode configuration
	Mode PaymentMode

	// Sandbox credentials
	SandboxAPIKey        string
	SandboxWebhookSecret string
	SandboxProductIDPro  string // Pro subscription product ID

	// Production credentials
	LiveAPIKey        string
	LiveWebhookSecret string
	LiveProductIDPro  string // Pro subscription product ID

	// Frontend URLs
	CheckoutSuccessURL string
	CheckoutCancelURL  string

	// Webhook configuration
	WebhookToleranceSeconds int

	// Logger
	Logger *zap.Logger
}

// NewPaddleConfig creates a new Paddle configuration from environment variables
func NewPaddleConfig(logger *zap.Logger) (*PaddleConfig, error) {
	mode := getEnvOrDefault("PAYMENT_MODE", "sandbox")
	paymentMode := PaymentMode(strings.ToLower(mode))

	// Validate payment mode
	if paymentMode != PaymentModeMock && paymentMode != PaymentModeSandbox && paymentMode != PaymentModeProduction {
		return nil, fmt.Errorf("invalid PAYMENT_MODE: %s (must be mock, sandbox, or production)", mode)
	}

	config := &PaddleConfig{
		Mode:   paymentMode,
		Logger: logger,

		// Sandbox credentials
		SandboxAPIKey:       os.Getenv("PADDLE_SANDBOX_API_KEY"),
		SandboxWebhookSecret: os.Getenv("PADDLE_SANDBOX_WEBHOOK_SECRET"),
		SandboxProductIDPro: os.Getenv("PADDLE_SANDBOX_PRODUCT_ID_PRO"),

		// Production credentials
		LiveAPIKey:       os.Getenv("PADDLE_LIVE_API_KEY"),
		LiveWebhookSecret: os.Getenv("PADDLE_LIVE_WEBHOOK_SECRET"),
		LiveProductIDPro: os.Getenv("PADDLE_LIVE_PRODUCT_ID_PRO"),

		// Frontend URLs
		CheckoutSuccessURL: getEnvOrDefault("FRONTEND_CHECKOUT_SUCCESS_URL", "http://localhost:3000/subscription-success"),
		CheckoutCancelURL:  getEnvOrDefault("FRONTEND_CHECKOUT_CANCEL_URL", "http://localhost:3000/pricing"),

		// Webhook configuration
		WebhookToleranceSeconds: getEnvAsIntOrDefault("PADDLE_WEBHOOK_TOLERANCE_SECONDS", 300),
	}

	// Validate configuration based on mode
	if err := config.validate(); err != nil {
		return nil, err
	}

	logger.Info("Paddle configuration initialized",
		zap.String("mode", string(paymentMode)),
		zap.Bool("sandbox_configured", config.SandboxAPIKey != ""),
		zap.Bool("production_configured", config.LiveAPIKey != ""),
	)

	return config, nil
}

// validate ensures required configuration is present based on payment mode
func (c *PaddleConfig) validate() error {
	switch c.Mode {
	case PaymentModeSandbox:
		if c.SandboxAPIKey == "" {
			return fmt.Errorf("PADDLE_SANDBOX_API_KEY is required for sandbox mode")
		}
		if c.SandboxWebhookSecret == "" {
			return fmt.Errorf("PADDLE_SANDBOX_WEBHOOK_SECRET is required for sandbox mode")
		}
		if c.SandboxProductIDPro == "" {
			return fmt.Errorf("PADDLE_SANDBOX_PRODUCT_ID_PRO is required for sandbox mode")
		}

	case PaymentModeProduction:
		if c.LiveAPIKey == "" {
			return fmt.Errorf("PADDLE_LIVE_API_KEY is required for production mode")
		}
		if c.LiveWebhookSecret == "" {
			return fmt.Errorf("PADDLE_LIVE_WEBHOOK_SECRET is required for production mode")
		}
		if c.LiveProductIDPro == "" {
			return fmt.Errorf("PADDLE_LIVE_PRODUCT_ID_PRO is required for production mode")
		}

	case PaymentModeMock:
		// Mock mode doesn't require any credentials
		c.Logger.Info("Running in mock payment mode - no real Paddle calls will be made")
	}

	return nil
}

// GetAPIKey returns the appropriate API key based on payment mode
func (c *PaddleConfig) GetAPIKey() string {
	switch c.Mode {
	case PaymentModeSandbox:
		return c.SandboxAPIKey
	case PaymentModeProduction:
		return c.LiveAPIKey
	default:
		return ""
	}
}

// GetWebhookSecret returns the appropriate webhook secret based on payment mode
func (c *PaddleConfig) GetWebhookSecret() string {
	switch c.Mode {
	case PaymentModeSandbox:
		return c.SandboxWebhookSecret
	case PaymentModeProduction:
		return c.LiveWebhookSecret
	default:
		return ""
	}
}

// GetProductID returns the Paddle product ID for the Pro tier
func (c *PaddleConfig) GetProductID(tier string) (string, error) {
	// Only Pro tier is supported
	tier = strings.ToLower(tier)
	if tier != "pro" {
		return "", fmt.Errorf("only 'pro' tier is supported, got: %s", tier)
	}

	var productID string
	switch c.Mode {
	case PaymentModeSandbox:
		productID = c.SandboxProductIDPro
	case PaymentModeProduction:
		productID = c.LiveProductIDPro
	case PaymentModeMock:
		// Return mock product ID
		return "mock_product_pro", nil
	}

	if productID == "" {
		return "", fmt.Errorf("Pro product ID not configured for mode: %s", c.Mode)
	}

	return productID, nil
}

// IsMockMode returns true if running in mock mode
func (c *PaddleConfig) IsMockMode() bool {
	return c.Mode == PaymentModeMock
}

// IsSandboxMode returns true if running in sandbox mode
func (c *PaddleConfig) IsSandboxMode() bool {
	return c.Mode == PaymentModeSandbox
}

// IsProductionMode returns true if running in production mode
func (c *PaddleConfig) IsProductionMode() bool {
	return c.Mode == PaymentModeProduction
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
