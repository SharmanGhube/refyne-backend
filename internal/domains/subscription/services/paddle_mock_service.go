package services

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/config"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

// MockPaddleService implements PaddleService for mock/testing mode
type MockPaddleService struct {
	name   string
	config *config.PaddleConfig
	logger *zap.Logger
}

// NewMockPaddleService creates a new mock Paddle service
func NewMockPaddleService(cfg *config.PaddleConfig) PaddleService {
	return &MockPaddleService{
		name:   "MockPaddleService",
		config: cfg,
		logger: cfg.Logger,
	}
}

// GenerateCheckoutURL returns a mock checkout URL
func (m *MockPaddleService) GenerateCheckoutURL(ctx *gin.Context, userID, userEmail, tier string) (string, *errors.AppError) {
	// In mock mode, redirect directly to success page with mock parameter
	checkoutURL := fmt.Sprintf("%s?mock=true&tier=%s&user_id=%s",
		m.config.CheckoutSuccessURL,
		tier,
		userID,
	)

	m.logger.Info("Generated mock checkout URL",
		zap.String("user_id", userID),
		zap.String("tier", tier),
	)

	return checkoutURL, nil
}

// GetCustomerPortalURL returns a mock portal URL
func (m *MockPaddleService) GetCustomerPortalURL(ctx *gin.Context, customerID string) (string, *errors.AppError) {
	portalURL := fmt.Sprintf("%s?mock=true&customer_id=%s",
		m.config.CheckoutSuccessURL,
		customerID,
	)

	m.logger.Info("Generated mock customer portal URL",
		zap.String("customer_id", customerID),
	)

	return portalURL, nil
}

// VerifyWebhookSignature always returns true in mock mode
func (m *MockPaddleService) VerifyWebhookSignature(payload []byte, signature string) bool {
	m.logger.Info("Mock webhook signature verification (always true)")
	return true
}

// GetMode returns "mock"
func (m *MockPaddleService) GetMode() string {
	return "mock"
}
