package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

// WebhookDeduplicator handles deduplication of webhook events using Redis
type WebhookDeduplicator interface {
	// Check if event has been processed recently (24h window)
	IsProcessed(ctx context.Context, eventID string) (bool, error)

	// Mark event as processed with 24h TTL
	MarkProcessed(ctx context.Context, eventID string) error

	// Get deduplication stats
	GetStats(ctx context.Context) map[string]interface{}
}

type redisWebhookDeduplicator struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewWebhookDeduplicator creates a new Redis-backed deduplicator
func NewWebhookDeduplicator(redisClient *redis.Client) WebhookDeduplicator {
	return &redisWebhookDeduplicator{
		redis:  redisClient,
		logger: logging.GetServiceLogger("WebhookDeduplicator"),
	}
}

// IsProcessed checks if an event has been processed in the deduplication window
func (d *redisWebhookDeduplicator) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	key := d.getDedupKey(eventID)
	val, err := d.redis.Get(ctx, key).Result()

	if err == redis.Nil {
		// Not found - not processed
		return false, nil
	}

	if err != nil {
		d.logger.Error("Failed to check dedup status", zap.Error(err), zap.String("event_id", eventID))
		return false, err
	}

	// Value exists - already processed
	return val == "1", nil
}

// MarkProcessed marks an event as processed with 24h TTL
func (d *redisWebhookDeduplicator) MarkProcessed(ctx context.Context, eventID string) error {
	key := d.getDedupKey(eventID)
	ttl := 24 * time.Hour

	err := d.redis.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		d.logger.Error("Failed to mark event as processed",
			zap.Error(err),
			zap.String("event_id", eventID),
		)
		return err
	}

	return nil
}

// GetStats returns deduplication statistics
func (d *redisWebhookDeduplicator) GetStats(ctx context.Context) map[string]interface{} {
	// Scan for all webhook dedup keys and count them
	iter := d.redis.Scan(ctx, 0, "instagram:webhook:*", 0).Iterator()
	count := 0

	for iter.Next(ctx) {
		count++
	}

	return map[string]interface{}{
		"deduplicated_events": count,
		"prefix":              "instagram:webhook:",
		"ttl":                 "24h",
	}
}

// getDedupKey returns the Redis key for a webhook event
func (d *redisWebhookDeduplicator) getDedupKey(eventID string) string {
	// Use SHA256 hash of event ID to ensure consistent length
	hash := sha256.Sum256([]byte(eventID))
	hashStr := hex.EncodeToString(hash[:])

	return "instagram:webhook:" + hashStr
}

// RateLimitChecker handles Instagram API rate limiting
type RateLimitChecker interface {
	// Check if account can make API calls (200/hour limit)
	CanMakeCall(ctx context.Context, accountID string) (bool, int, error)

	// Record an API call
	RecordCall(ctx context.Context, accountID string) error

	// Get rate limit status
	GetStatus(ctx context.Context, accountID string) map[string]interface{}
}

type redisRateLimitChecker struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewRateLimitChecker creates a rate limit checker
func NewRateLimitChecker(redisClient *redis.Client) RateLimitChecker {
	return &redisRateLimitChecker{
		redis:  redisClient,
		logger: logging.GetServiceLogger("RateLimitChecker"),
	}
}

const (
	maxCallsPerHour = 200
	safeThreshold   = 180 // Keep 20 as safety margin
	windowSize      = time.Hour
)

// CanMakeCall checks if an account can make API calls
func (r *redisRateLimitChecker) CanMakeCall(ctx context.Context, accountID string) (bool, int, error) {
	key := "instagram:ratelimit:" + accountID

	// Get number of calls in current window
	count, err := r.redis.ZCard(ctx, key).Result()
	if err != nil {
		r.logger.Error("Failed to check rate limit", zap.Error(err), zap.String("account_id", accountID))
		return false, 0, err
	}

	remaining := safeThreshold - int(count)

	if count >= int64(safeThreshold) {
		r.logger.Warn("Rate limit threshold reached",
			zap.String("account_id", accountID),
			zap.Int64("calls", count),
		)
		return false, 0, nil
	}

	return true, remaining, nil
}

// RecordCall records an API call in the rate limit window
func (r *redisRateLimitChecker) RecordCall(ctx context.Context, accountID string) error {
	key := "instagram:ratelimit:" + accountID
	now := time.Now()

	// Add to sorted set with timestamp as score
	err := r.redis.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.Unix()),
		Member: now.Format(time.RFC3339Nano),
	}).Err()

	if err != nil {
		r.logger.Error("Failed to record API call", zap.Error(err), zap.String("account_id", accountID))
		return err
	}

	// Remove old entries and set expiry
	cutoff := float64(now.Add(-windowSize).Unix())
	r.redis.ZRemRangeByScore(ctx, key, "-inf", fmt.Sprintf("%f", cutoff))
	r.redis.Expire(ctx, key, windowSize)

	return nil
}

// GetStatus returns rate limit status for an account
func (r *redisRateLimitChecker) GetStatus(ctx context.Context, accountID string) map[string]interface{} {
	key := "instagram:ratelimit:" + accountID

	count, err := r.redis.ZCard(ctx, key).Result()
	if err != nil {
		r.logger.Warn("Failed to get rate limit status", zap.Error(err))
		count = 0
	}

	remaining := safeThreshold - int(count)
	if remaining < 0 {
		remaining = 0
	}

	return map[string]interface{}{
		"account_id":      accountID,
		"calls_used":      int(count),
		"calls_remaining": remaining,
		"max_per_hour":    maxCallsPerHour,
		"safe_threshold":  safeThreshold,
	}
}
