package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/config"
	errors "github.com/refynehq/refyne-backend/pkg/error"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// WebhookEvent represents an incoming Instagram webhook event
type WebhookEvent struct {
	Object string       `json:"object"`
	Entry  []EventEntry `json:"entry"`
}

type EventEntry struct {
	ID      string        `json:"id"`
	Time    int64         `json:"time"`
	Changes []EventChange `json:"changes"`
}

type EventChange struct {
	Value json.RawMessage `json:"value"`
	Field string          `json:"field"`
}

// InstagramWebhookService handles webhook verification and processing
type InstagramWebhookService interface {
	// Verify incoming webhook signature (challenge or event)
	VerifyWebhookSignature(body []byte, signature string) bool

	// Parse webhook event
	ParseWebhookEvent(body []byte) (*WebhookEvent, *errors.AppError)

	// Validate webhook timestamp (prevent replay attacks)
	ValidateWebhookTimestamp(timestamp int64) bool
}

type instagramWebhookService struct {
	config *config.InstagramConfig
	logger *zap.Logger
}

// NewInstagramWebhookService creates a new webhook service
func NewInstagramWebhookService(cfg *config.InstagramConfig) InstagramWebhookService {
	return &instagramWebhookService{
		config: cfg,
		logger: logging.GetServiceLogger("InstagramWebhookService"),
	}
}

// VerifyWebhookSignature verifies the HMAC-SHA256 signature of the webhook
// Instagram sends: X-Hub-Signature-256: sha256=<signature>
func (s *instagramWebhookService) VerifyWebhookSignature(body []byte, signature string) bool {
	if s.config == nil || s.config.AppSecret == "" {
		s.logger.Warn("Instagram app secret not configured, signature verification skipped")
		return false
	}

	// Create HMAC-SHA256 of the request body
	h := hmac.New(sha256.New, []byte(s.config.AppSecret))
	h.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	// Compare signatures using constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// ParseWebhookEvent parses the JSON webhook payload
func (s *instagramWebhookService) ParseWebhookEvent(body []byte) (*WebhookEvent, *errors.AppError) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		s.logger.Error("Failed to parse webhook event", zap.Error(err))
		return nil, errors.NewAppError(
			&gin.Context{},
			"WEBHOOK_PARSE_ERROR",
			"Failed to parse webhook event",
			errors.ErrorTypeValidation,
			errors.SeverityMedium,
			"instagram",
		).WithContext("error", err.Error())
	}

	return &event, nil
}

// ValidateWebhookTimestamp ensures the webhook is recent (within 5 minutes)
// This prevents replay attacks
func (s *instagramWebhookService) ValidateWebhookTimestamp(timestamp int64) bool {
	if timestamp == 0 {
		return false
	}

	eventTime := time.Unix(timestamp, 0)
	now := time.Now()
	timeDiff := now.Sub(eventTime)

	// Allow 5 minute window for webhook delivery
	if timeDiff < 0 || timeDiff > 5*time.Minute {
		s.logger.Warn("Webhook timestamp outside acceptable window",
			zap.Int64("timestamp", timestamp),
			zap.Duration("age", timeDiff),
		)
		return false
	}

	return true
}
