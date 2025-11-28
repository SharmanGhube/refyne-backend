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

// CreateCheckout generates a Paddle checkout URL
func (h *SubscriptionHandlerImpl) CreateCheckout(c *gin.Context) {
	// Extract user from JWT
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userEmail, _ := middlewares.GetUserEmail(c)

	// Parse request
	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid checkout request",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Generate checkout URL
	checkoutURL, appErr := h.paddleService.GenerateCheckoutURL(c, userID, userEmail, req.Tier)
	if appErr != nil {
		h.logger.Error("Failed to generate checkout URL",
			zap.String("user_id", userID),
			zap.String("tier", req.Tier),
		)
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	response := models.CheckoutResponse{
		CheckoutURL: checkoutURL,
		Mode:        h.paddleService.GetMode(),
	}

	h.logger.Info("Checkout URL generated",
		zap.String("user_id", userID),
		zap.String("tier", req.Tier),
	)

	c.JSON(http.StatusOK, response)
}

// GetSubscriptionStatus returns user's subscription details
func (h *SubscriptionHandlerImpl) GetSubscriptionStatus(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get subscription status
	tier, status, expiresAt, customerID, _, appErr := h.subscriptionRepo.GetUserSubscriptionStatus(c, userID)
	if appErr != nil {
		h.logger.Error("Failed to get subscription status",
			zap.String("user_id", userID),
		)
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	// Generate customer portal URL if customer ID exists
	var portalURL string
	if customerID != nil && *customerID != "" {
		url, _ := h.paddleService.GetCustomerPortalURL(c, *customerID)
		portalURL = url
	}

	// Determine if user can upgrade/downgrade
	canUpgrade := tier != "enterprise" && status == "active"
	canDowngrade := tier != "starter" && status == "active"

	response := models.SubscriptionStatusResponse{
		Tier:                tier,
		Status:              status,
		ExpiresAt:           expiresAt,
		PaddleCustomerID:    customerID,
		CanUpgrade:          canUpgrade,
		CanDowngrade:        canDowngrade,
		ManagementPortalURL: portalURL,
	}

	c.JSON(http.StatusOK, response)
}

// GetCustomerPortal generates customer portal URL
func (h *SubscriptionHandlerImpl) GetCustomerPortal(c *gin.Context) {
	userID, exists := middlewares.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user's Paddle customer ID
	_, _, _, customerID, _, appErr := h.subscriptionRepo.GetUserSubscriptionStatus(c, userID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	if customerID == nil || *customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No active subscription found",
		})
		return
	}

	// Generate portal URL
	portalURL, appErr := h.paddleService.GetCustomerPortalURL(c, *customerID)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	response := models.CustomerPortalResponse{
		PortalURL: portalURL,
	}

	c.JSON(http.StatusOK, response)
}

// HandleWebhook processes Paddle webhook events
func (h *SubscriptionHandlerImpl) HandleWebhook(c *gin.Context) {
	// Read raw body for signature verification
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read webhook body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Verify webhook signature (if not in mock mode)
	signature := c.GetHeader("Paddle-Signature")
	if !h.paddleService.VerifyWebhookSignature(bodyBytes, signature) {
		h.logger.Warn("Invalid webhook signature",
			zap.String("signature", signature),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse webhook event
	var event models.PaddleWebhookEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("Failed to parse webhook event", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook format"})
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
		c.JSON(appErr.HTTPStatus, gin.H{"error": appErr.Message})
		return
	}

	// Return 200 to acknowledge webhook
	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"event_id": event.EventID,
	})
}
