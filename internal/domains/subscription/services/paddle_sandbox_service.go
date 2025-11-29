package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/PaddleHQ/paddle-go-sdk/v3"
	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/config"
	subscriptionErrors "github.com/refynehq/refyne-backend/internal/domains/subscription/errors"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"go.uber.org/zap"
)

// PaddleSandboxService implements PaddleService for sandbox/production mode
type PaddleSandboxService struct {
	name   string
	config *config.PaddleConfig
	client *paddle.SDK
	logger *zap.Logger
}

// NewPaddleSandboxService creates a new Paddle service for sandbox mode
func NewPaddleSandboxService(cfg *config.PaddleConfig) (PaddleService, error) {
	apiKey := cfg.GetAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("paddle API key is required for sandbox/production mode")
	}

	// Initialize Paddle SDK client
	var client *paddle.SDK
	var err error

	if cfg.IsSandboxMode() {
		client, err = paddle.NewSandbox(apiKey)
	} else {
		client, err = paddle.New(apiKey)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to initialize Paddle client: %w", err)
	}

	return &PaddleSandboxService{
		name:   "PaddleSandboxService",
		config: cfg,
		client: client,
		logger: cfg.Logger,
	}, nil
}

// GenerateCheckoutURL creates a Paddle checkout session using Paddle Billing API v3
func (s *PaddleSandboxService) GenerateCheckoutURL(ctx *gin.Context, userID, userEmail, tier string) (string, *errors.AppError) {
	// Get the price ID (Paddle Billing v3 uses price IDs, not product IDs)
	priceID, err := s.config.GetProductID(tier)
	if err != nil {
		s.logger.Error("Failed to get price ID",
			zap.String("tier", tier),
			zap.Error(err),
		)
		return "", subscriptionErrors.NewInvalidSubscriptionTierError(ctx, tier)
	}

	s.logger.Info("Creating Paddle transaction",
		zap.String("user_id", userID),
		zap.String("tier", tier),
		zap.String("price_id", priceID),
		zap.String("mode", s.GetMode()),
	)

	// Create transaction using Paddle SDK v3
	// Build transaction items using catalog price
	catalogItem := &paddle.TransactionItemFromCatalog{
		PriceID:  priceID,
		Quantity: 1,
	}

	// Wrap in CreateTransactionItems union type
	items := []paddle.CreateTransactionItems{
		*paddle.NewCreateTransactionItemsTransactionItemFromCatalog(catalogItem),
	}

	// Build custom data
	customData := paddle.CustomData{
		"user_id": userID,
		"tier":    tier,
	}

	// Create transaction request
	// NOTE: Do NOT set Checkout.URL - Paddle requires a default checkout URL
	// to be configured in the dashboard. The Checkout.URL field is for overriding
	// that default, not for setting success redirect URLs.
	transactionReq := &paddle.CreateTransactionRequest{
		Items:      items,
		CustomData: customData,
	}

	s.logger.Info("Calling Paddle API to create transaction",
		zap.String("price_id", priceID),
		zap.String("customer_email", userEmail),
	)

	// Create the transaction via Paddle API
	transaction, err := s.client.CreateTransaction(ctx.Request.Context(), transactionReq)
	if err != nil {
		s.logger.Error("Failed to create Paddle transaction",
			zap.String("user_id", userID),
			zap.String("tier", tier),
			zap.Error(err),
		)
		return "", subscriptionErrors.NewCheckoutCreationError(ctx, err)
	}

	// Extract checkout URL from transaction
	checkoutURL := ""
	if transaction != nil && transaction.Checkout != nil && transaction.Checkout.URL != nil {
		checkoutURL = *transaction.Checkout.URL
	} else {
		s.logger.Error("No checkout URL in Paddle transaction response",
			zap.String("transaction_id", transaction.ID),
		)
		return "", subscriptionErrors.NewCheckoutCreationError(ctx, fmt.Errorf("no checkout URL returned from Paddle"))
	}

	s.logger.Info("Paddle transaction created successfully",
		zap.String("transaction_id", transaction.ID),
		zap.String("checkout_url", checkoutURL),
		zap.String("user_id", userID),
	)

	return checkoutURL, nil
}

// GetCustomerPortalURL generates customer portal URL
func (s *PaddleSandboxService) GetCustomerPortalURL(ctx *gin.Context, customerID string) (string, *errors.AppError) {
	// Paddle customer portal URL format
	portalURL := fmt.Sprintf("https://checkout.paddle.com/subscription/update/%s", customerID)

	if s.config.IsSandboxMode() {
		portalURL = fmt.Sprintf("https://sandbox-checkout.paddle.com/subscription/update/%s", customerID)
	}

	s.logger.Info("Generated customer portal URL",
		zap.String("customer_id", customerID),
	)

	return portalURL, nil
}

// VerifyWebhookSignature validates webhook signature from Paddle
func (s *PaddleSandboxService) VerifyWebhookSignature(payload []byte, signature string) bool {
	webhookSecret := s.config.GetWebhookSecret()
	if webhookSecret == "" {
		s.logger.Warn("Webhook secret not configured")
		return false
	}

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	isValid := hmac.Equal([]byte(signature), []byte(expectedMAC))

	if !isValid {
		s.logger.Warn("Invalid webhook signature")
	}

	return isValid
}

// GetMode returns the current payment mode
func (s *PaddleSandboxService) GetMode() string {
	return string(s.config.Mode)
}
