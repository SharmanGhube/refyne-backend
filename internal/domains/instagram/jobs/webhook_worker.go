package jobs

import (
	"context"
	"encoding/json"

	"github.com/refynehq/refyne-backend/pkg/logging"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// InstagramWebhookArgs represents arguments for webhook processing job
type InstagramWebhookArgs struct {
	EventID   string      `json:"event_id"`
	Object    string      `json:"object"`
	Entry     []EntryData `json:"entry"`
	Timestamp int64       `json:"timestamp"`
}

type EntryData struct {
	ID      string       `json:"id"`
	Time    int64        `json:"time"`
	Changes []ChangeData `json:"changes"`
}

type ChangeData struct {
	Value json.RawMessage `json:"value"`
	Field string          `json:"field"`
}

func (InstagramWebhookArgs) Kind() string { return "instagram_webhook" }

// InstagramWebhookWorker processes incoming Instagram webhook events
type InstagramWebhookWorker struct {
	river.WorkerDefaults[InstagramWebhookArgs]
	logger *zap.Logger
}

// NewInstagramWebhookWorker creates a new webhook worker
func NewInstagramWebhookWorker() *InstagramWebhookWorker {
	return &InstagramWebhookWorker{
		logger: logging.GetJobLogger("InstagramWebhookWorker"),
	}
}

// Work processes the webhook event
// This job handles deduplication, field routing, and downstream processing
func (w *InstagramWebhookWorker) Work(ctx context.Context, job *river.Job[InstagramWebhookArgs]) error {
	w.logger.Info("Processing Instagram webhook",
		zap.String("event_id", job.Args.EventID),
		zap.String("object", job.Args.Object),
		zap.Int("entries", len(job.Args.Entry)),
	)

	// TODO: Deduplicate using Redis
	// Check if this event_id has been processed in the last 24h
	// If yes, return early

	// Process each entry in the payload
	for _, entry := range job.Args.Entry {
		if err := w.processEntry(ctx, entry); err != nil {
			w.logger.Error("Failed to process entry",
				zap.String("user_id", entry.ID),
				zap.Error(err),
			)
			// Don't return error - continue processing other entries
		}
	}

	w.logger.Info("Webhook processing complete", zap.String("event_id", job.Args.EventID))
	return nil
}

// processEntry routes the entry to the appropriate handler based on change field
func (w *InstagramWebhookWorker) processEntry(ctx context.Context, entry EntryData) error {
	for _, change := range entry.Changes {
		switch change.Field {
		case "feed":
			// Media changed in feed (new post, edit, delete)
			return w.processFeedChange(ctx, change.Value, entry.ID)

		case "story":
			// Story-related event
			return w.processStoryChange(ctx, change.Value, entry.ID)

		case "messages":
			// Direct message received
			return w.processMessageChange(ctx, change.Value, entry.ID)

		case "message_template_status_update":
			// Message template status changed
			w.logger.Debug("Message template status update", zap.String("user_id", entry.ID))

		case "message_template_quality_update":
			// Message template quality changed
			w.logger.Debug("Message template quality update", zap.String("user_id", entry.ID))

		default:
			w.logger.Debug("Unknown webhook field", zap.String("field", change.Field))
		}
	}

	return nil
}

// processFeedChange handles feed changes (posts, edits, deletes)
func (w *InstagramWebhookWorker) processFeedChange(ctx context.Context, changeValue json.RawMessage, accountID string) error {
	var feedChange struct {
		MediaID string `json:"media_id"`
		Status  string `json:"status"` // PUBLISHED, EXPIRED, etc
		From    string `json:"from"`
	}

	if err := json.Unmarshal(changeValue, &feedChange); err != nil {
		w.logger.Warn("Failed to parse feed change", zap.Error(err))
		return nil // Don't error on parse failure
	}

	w.logger.Info("Feed change detected",
		zap.String("account_id", accountID),
		zap.String("media_id", feedChange.MediaID),
		zap.String("status", feedChange.Status),
	)

	// TODO: Queue media sync job
	// This will fetch latest media and update cache

	return nil
}

// processStoryChange handles story updates
func (w *InstagramWebhookWorker) processStoryChange(ctx context.Context, changeValue json.RawMessage, accountID string) error {
	// Stories have 24h expiry, less critical than posts
	w.logger.Debug("Story change detected", zap.String("account_id", accountID))

	// TODO: Handle story updates if needed

	return nil
}

// processMessageChange handles direct messages
func (w *InstagramWebhookWorker) processMessageChange(ctx context.Context, changeValue json.RawMessage, accountID string) error {
	var messageChange struct {
		From      string `json:"from"`
		To        string `json:"to"`
		MessageID string `json:"message_id"`
		Text      string `json:"text"`
	}

	if err := json.Unmarshal(changeValue, &messageChange); err != nil {
		w.logger.Warn("Failed to parse message change", zap.Error(err))
		return nil
	}

	w.logger.Info("Direct message received",
		zap.String("account_id", accountID),
		zap.String("from", messageChange.From),
		zap.String("message_id", messageChange.MessageID),
	)

	// TODO: Store message in database
	// TODO: Notify user if configured
	// TODO: Auto-respond if configured

	return nil
}
