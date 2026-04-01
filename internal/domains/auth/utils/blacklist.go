package auth

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// globalBlacklistManager holds the active blacklist manager (either Redis or in-memory).
var (
	globalBlacklistManager     TokenBlacklistManager
	globalBlacklistManagerOnce sync.Once
	globalBlacklistManagerMu   sync.RWMutex
)

// LegacyTokenBlacklistManager provides backward-compatible API.
// It wraps the new TokenBlacklistManager interface.
type LegacyTokenBlacklistManager struct {
	manager TokenBlacklistManager
}

// InitTokenBlacklistManager initializes the global token blacklist manager with Redis.
// Call this during application startup with a Redis client.
// If redisClient is nil, falls back to in-memory implementation.
func InitTokenBlacklistManager(redisClient *redis.Client) {
	globalBlacklistManagerMu.Lock()
	defer globalBlacklistManagerMu.Unlock()

	if redisClient != nil {
		globalBlacklistManager = NewRedisTokenBlacklistManager(redisClient)
	} else {
		globalBlacklistManager = GetInMemoryTokenBlacklistManager()
	}
}

// GetTokenBlacklistManager returns the backward-compatible blacklist manager.
// This maintains the existing API signature for existing code.
func GetTokenBlacklistManager() *LegacyTokenBlacklistManager {
	globalBlacklistManagerOnce.Do(func() {
		globalBlacklistManagerMu.Lock()
		defer globalBlacklistManagerMu.Unlock()
		if globalBlacklistManager == nil {
			// Default to in-memory if not initialized
			globalBlacklistManager = GetInMemoryTokenBlacklistManager()
		}
	})

	globalBlacklistManagerMu.RLock()
	defer globalBlacklistManagerMu.RUnlock()

	return &LegacyTokenBlacklistManager{
		manager: globalBlacklistManager,
	}
}

// BlacklistToken adds a token to the blacklist (backward-compatible API).
func (l *LegacyTokenBlacklistManager) BlacklistToken(token string, expiresAt time.Time, reason string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Ignore error for backward compatibility
	_ = l.manager.BlacklistToken(ctx, token, expiresAt, reason)
}

// IsBlacklisted checks if a token is blacklisted (backward-compatible API).
func (l *LegacyTokenBlacklistManager) IsBlacklisted(token string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result, err := l.manager.IsBlacklisted(ctx, token)
	if err != nil {
		// On error, assume not blacklisted to avoid blocking users
		return false
	}
	return result
}

// RemoveToken removes a token from the blacklist.
func (l *LegacyTokenBlacklistManager) RemoveToken(token string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = l.manager.RemoveToken(ctx, token)
}

// GetBlacklistedCount returns the number of blacklisted tokens.
func (l *LegacyTokenBlacklistManager) GetBlacklistedCount() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	count, err := l.manager.GetBlacklistedCount(ctx)
	if err != nil {
		return 0
	}
	return count
}

// GetTokenInfo returns information about a blacklisted token.
func (l *LegacyTokenBlacklistManager) GetTokenInfo(token string) (*TokenBlacklistEntry, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	entry, exists, err := l.manager.GetTokenInfo(ctx, token)
	if err != nil {
		return nil, false
	}
	return entry, exists
}

// ClearAll removes all tokens from blacklist.
func (l *LegacyTokenBlacklistManager) ClearAll() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = l.manager.ClearAll(ctx)
}

// GetUnderlyingManager returns the underlying TokenBlacklistManager.
// Use this for new code that wants the full interface with context and error handling.
func (l *LegacyTokenBlacklistManager) GetUnderlyingManager() TokenBlacklistManager {
	return l.manager
}
