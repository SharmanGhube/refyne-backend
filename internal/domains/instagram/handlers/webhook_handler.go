package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/jobs"
	"go.uber.org/zap"
)

// HandleWebhook processes incoming Instagram webhooks
// GET /api/instagram/webhooks?hub.mode=subscribe&hub.challenge=...&hub.verify_token=...
// POST /api/instagram/webhooks (with X-Hub-Signature-256 header)
func (h *InstagramHandler) HandleWebhook(c *gin.Context) {
	// Verify token for subscription challenge
	if c.Query("hub.mode") == "subscribe" {
		challenge := c.Query("hub.challenge")

		// In production, verify_token should come from Instagram config
		// For now, we'll skip it since it's a setup parameter
		if challenge != "" {
			h.logger.Info("Webhook challenge received", zap.String("challenge", challenge))
			c.String(200, challenge)
			return
		}

		c.JSON(400, gin.H{"error": "Missing challenge parameter"})
		return
	}

	// Handle incoming webhook events
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Warn("Failed to read webhook body", zap.Error(err))
		c.JSON(400, gin.H{"error": "Failed to read body"})
		return
	}

	// Verify signature
	signature := c.GetHeader("X-Hub-Signature-256")
	if signature == "" {
		h.logger.Warn("Missing webhook signature")
		c.JSON(403, gin.H{"error": "Missing signature"})
		return
	}

	if !h.webhookService.VerifyWebhookSignature(body, signature) {
		h.logger.Warn("Invalid webhook signature", zap.String("signature", signature[:20]+"..."))
		c.JSON(403, gin.H{"error": "Invalid signature"})
		return
	}

	// Parse event
	event, appErr := h.webhookService.ParseWebhookEvent(body)
	if appErr != nil {
		c.JSON(400, gin.H{"error": appErr.Message})
		return
	}

	// Validate timestamp
	for _, entry := range event.Entry {
		if !h.webhookService.ValidateWebhookTimestamp(entry.Time) {
			h.logger.Warn("Webhook timestamp outside acceptable window",
				zap.Int64("timestamp", entry.Time),
			)
			// Still acknowledge receipt
			c.JSON(200, gin.H{"status": "ok"})
			return
		}
	}

	// Generate event ID for deduplication (hash of body)
	hash := sha256.Sum256(body)
	eventID := hex.EncodeToString(hash[:])

	// Check if event has already been processed
	processed, err := h.deduplicator.IsProcessed(c.Request.Context(), eventID)
	if err != nil {
		h.logger.Error("Failed to check dedup status", zap.Error(err), zap.String("event_id", eventID))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	if processed {
		h.logger.Info("Webhook event already processed, skipping", zap.String("event_id", eventID))
		c.JSON(200, gin.H{"status": "ok"})
		return
	}

	// Mark event as processed
	if err := h.deduplicator.MarkProcessed(c.Request.Context(), eventID); err != nil {
		h.logger.Error("Failed to mark event as processed", zap.Error(err), zap.String("event_id", eventID))
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	// Queue webhook processing job via River
	riverClient := h.riverService.GetClient()

	// Convert entries to job format
	entries := make([]jobs.EntryData, len(event.Entry))
	for i, entry := range event.Entry {
		changes := make([]jobs.ChangeData, len(entry.Changes))
		for j, change := range entry.Changes {
			changes[j] = jobs.ChangeData{
				Field: change.Field,
				Value: change.Value,
			}
		}
		entries[i] = jobs.EntryData{
			ID:      entry.ID,
			Time:    entry.Time,
			Changes: changes,
		}
	}

	jobArgs := jobs.InstagramWebhookArgs{
		EventID:   eventID,
		Object:    event.Object,
		Entry:     entries,
		Timestamp: h.getCurrentTimestamp(),
	}

	_, err = riverClient.Insert(context.Background(), jobArgs, nil)
	if err != nil {
		h.logger.Error("Failed to queue webhook job", zap.Error(err), zap.String("event_id", eventID))
		c.JSON(500, gin.H{"error": "Failed to process webhook"})
		return
	}

	h.logger.Info("Webhook event queued for processing",
		zap.String("event_id", eventID),
		zap.String("object", event.Object),
		zap.Int("entries", len(event.Entry)),
	)

	c.JSON(200, gin.H{"status": "ok"})
}

// getCurrentTimestamp returns current Unix timestamp
func (h *InstagramHandler) getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
