package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/api/middlewares"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/models"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/repository"
	"github.com/refynehq/refyne-backend/internal/domains/subscription/services"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type SubscriptionHandlerImpl struct {
	name             string
	logger           *zap.Logger
	paddleService    services.PaddleService
	webhookService   services.WebhookService
	subscriptionRepo repository.SubscriptionRepository
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(
	paddleService services.PaddleService,
	webhookService services.WebhookService,
	subscriptionRepo repository.SubscriptionRepository,
) SubscriptionHandler {
	return &SubscriptionHandlerImpl{
		name:             "SubscriptionHandler",
		logger:           logging.GetHandlerLogger("SubscriptionHandler"),
		paddleService:    paddleService,
		webhookService:   webhookService,
		subscriptionRepo: subscriptionRepo,
	}
}

// CreateCheckout generates a Paddle checkout URL for Pro subscription
func (h *SubscriptionHandlerImpl) CreateCheckout(c *gin.Context) {
	// Extract user from JWT
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	userEmail, _ := middlewares.GetUserEmail(c)

	// Parse request (tier parameter is optional/ignored for backwards compatibility)
	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		if err != io.EOF && err.Error() != "EOF" {
			h.logger.Warn("Invalid checkout request",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			middlewares.RespondWithError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request format", map[string]interface{}{
				"details": err.Error(),
			})
			return
		}
	}

	// Always use "pro" tier
	tier := "pro"

	// Generate checkout URL
	checkoutURL, appErr := h.paddleService.GenerateCheckoutURL(c, userID, userEmail, tier)
	if appErr != nil {
		h.logger.Error("Failed to generate checkout URL",
			zap.String("user_id", userID),
			zap.String("tier", tier),
		)
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	response := models.CheckoutResponse{
		CheckoutURL: checkoutURL,
		Mode:        h.paddleService.GetMode(),
	}

	h.logger.Info("Checkout URL generated",
		zap.String("user_id", userID),
		zap.String("tier", tier),
	)

	middlewares.RespondWithSuccess(c, http.StatusOK, "Checkout URL generated successfully", response)
}

// GetSubscriptionStatus returns user's subscription details
func (h *SubscriptionHandlerImpl) GetSubscriptionStatus(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	// Get subscription status
	tier, status, expiresAt, customerID, _, appErr := h.subscriptionRepo.GetUserSubscriptionStatus(c, userID)
	if appErr != nil {
		h.logger.Error("Failed to get subscription status",
			zap.String("user_id", userID),
		)
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	// Generate customer portal URL if customer ID exists
	var portalURL string
	if customerID != nil && *customerID != "" {
		url, _ := h.paddleService.GetCustomerPortalURL(c, *customerID)
		portalURL = url
	}

	response := models.SubscriptionStatusResponse{
		Tier:                tier,
		Status:              status,
		ExpiresAt:           expiresAt,
		PaddleCustomerID:    customerID,
		ManagementPortalURL: portalURL,
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Subscription status retrieved successfully", response)
}

// GetCustomerPortal generates customer portal URL
func (h *SubscriptionHandlerImpl) GetCustomerPortal(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		middlewares.RespondWithError(c, http.StatusUnauthorized, "UNAUTHORIZED", "User authentication required", nil)
		return
	}

	// Get user's Paddle customer ID
	_, _, _, customerID, _, appErr := h.subscriptionRepo.GetUserSubscriptionStatus(c, userID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	if customerID == nil || *customerID == "" {
		middlewares.RespondWithError(c, http.StatusBadRequest, "NO_SUBSCRIPTION", "No active subscription found", nil)
		return
	}

	// Generate portal URL
	portalURL, appErr := h.paddleService.GetCustomerPortalURL(c, *customerID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	response := models.CustomerPortalResponse{
		PortalURL: portalURL,
	}

	middlewares.RespondWithSuccess(c, http.StatusOK, "Customer portal URL generated successfully", response)
}

// HandleWebhook processes Paddle webhook events
func (h *SubscriptionHandlerImpl) HandleWebhook(c *gin.Context) {
	// Read raw body for signature verification
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Verify webhook signature (if not in mock mode)
	signature := c.GetHeader("Paddle-Signature")
	if !h.paddleService.VerifyWebhookSignature(bodyBytes, signature) {
		h.logger.Warn("Invalid webhook signature",
			zap.String("signature", signature),
		)
		middlewares.RespondWithError(c, http.StatusUnauthorized, "INVALID_SIGNATURE", "Invalid signature", nil)
		return
	}

	// Parse webhook event
	var event models.PaddleWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("Failed to parse webhook event", zap.Error(err))
		middlewares.RespondWithError(c, http.StatusBadRequest, "INVALID_FORMAT", "Invalid webhook format", nil)
		return
	}

	h.logger.Info("Received webhook",
		zap.String("event_id", event.EventID),
		zap.String("event_type", event.EventType),
	)

	// Process webhook
	appErr := h.webhookService.ProcessWebhook(c, &event)
	if appErr != nil {
		h.logger.Error("Failed to process webhook",
			zap.String("event_id", event.EventID),
			zap.String("event_type", event.EventType),
		)
		c.JSON(appErr.HTTPStatus, appErr.ClientResponse())
		return
	}

	// Return 200 to acknowledge webhook with standardized envelope
	middlewares.RespondWithSuccess(c, http.StatusOK, "Webhook processed successfully", gin.H{
		"event_id": event.EventID,
	})
}
