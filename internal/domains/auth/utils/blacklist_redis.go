package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBlacklistManager defines the interface for token blacklist operations.
// This allows for both in-memory and Redis implementations.
type TokenBlacklistManager interface {
	BlacklistToken(ctx context.Context, token string, expiresAt time.Time, reason string) error
	IsBlacklisted(ctx context.Context, token string) (bool, error)
	RemoveToken(ctx context.Context, token string) error
	GetBlacklistedCount(ctx context.Context) (int, error)
	GetTokenInfo(ctx context.Context, token string) (*TokenBlacklistEntry, bool, error)
	ClearAll(ctx context.Context) error
}

// TokenBlacklistEntry represents a blacklisted token with expiry.
type TokenBlacklistEntry struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	RevokedAt time.Time `json:"revoked_at"`
	Reason    string    `json:"reason"` // "logout", "password_reset", "security"
}

const (
	blacklistKeyPrefix = "token:blacklist:"
	blacklistSetKey    = "token:blacklist:set"
)

// RedisTokenBlacklistManager implements TokenBlacklistManager using Redis.
type RedisTokenBlacklistManager struct {
	client *redis.Client
}

// NewRedisTokenBlacklistManager creates a new Redis-backed token blacklist manager.
func NewRedisTokenBlacklistManager(client *redis.Client) *RedisTokenBlacklistManager {
	return &RedisTokenBlacklistManager{
		client: client,
	}
}

// BlacklistToken adds a token to the blacklist with automatic expiry.
func (m *RedisTokenBlacklistManager) BlacklistToken(ctx context.Context, token string, expiresAt time.Time, reason string) error {
	entry := TokenBlacklistEntry{
		Token:     token,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
		Reason:    reason,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal blacklist entry: %w", err)
	}

	key := blacklistKeyPrefix + token
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}

	// Set with TTL so Redis automatically removes expired entries
	if err := m.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// IsBlacklisted checks if a token is blacklisted.
func (m *RedisTokenBlacklistManager) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := blacklistKeyPrefix + token
	exists, err := m.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check blacklist: %w", err)
	}
	return exists > 0, nil
}

// RemoveToken removes a token from the blacklist.
func (m *RedisTokenBlacklistManager) RemoveToken(ctx context.Context, token string) error {
	key := blacklistKeyPrefix + token
	if err := m.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to remove token from blacklist: %w", err)
	}
	return nil
}

// GetBlacklistedCount returns the approximate count of blacklisted tokens.
func (m *RedisTokenBlacklistManager) GetBlacklistedCount(ctx context.Context) (int, error) {
	var cursor uint64
	var count int

	for {
		keys, nextCursor, err := m.client.Scan(ctx, cursor, blacklistKeyPrefix+"*", 100).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to scan blacklisted tokens: %w", err)
		}
		count += len(keys)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return count, nil
}

// GetTokenInfo returns information about a blacklisted token.
func (m *RedisTokenBlacklistManager) GetTokenInfo(ctx context.Context, token string) (*TokenBlacklistEntry, bool, error) {
	key := blacklistKeyPrefix + token
	data, err := m.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get token info: %w", err)
	}

	var entry TokenBlacklistEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal blacklist entry: %w", err)
	}

	return &entry, true, nil
}

// ClearAll removes all blacklisted tokens (for testing purposes).
func (m *RedisTokenBlacklistManager) ClearAll(ctx context.Context) error {
	var cursor uint64
	for {
		keys, nextCursor, err := m.client.Scan(ctx, cursor, blacklistKeyPrefix+"*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys for deletion: %w", err)
		}
		if len(keys) > 0 {
			if err := m.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("failed to delete keys: %w", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
