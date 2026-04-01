package auth

import (
	"context"
	"sync"
	"time"
)

// InMemoryTokenBlacklistManager is the in-memory implementation.
// For production with multiple instances, use RedisTokenBlacklistManager.
type InMemoryTokenBlacklistManager struct {
	tokens map[string]*TokenBlacklistEntry
	mu     sync.RWMutex
}

var (
	inMemoryBlacklistManager     *InMemoryTokenBlacklistManager
	inMemoryBlacklistManagerOnce sync.Once
)

// GetInMemoryTokenBlacklistManager returns the singleton in-memory blacklist manager.
// Deprecated: Use GetTokenBlacklistManagerWithRedis for production.
func GetInMemoryTokenBlacklistManager() *InMemoryTokenBlacklistManager {
	inMemoryBlacklistManagerOnce.Do(func() {
		inMemoryBlacklistManager = &InMemoryTokenBlacklistManager{
			tokens: make(map[string]*TokenBlacklistEntry),
		}
		go inMemoryBlacklistManager.cleanup()
	})
	return inMemoryBlacklistManager
}

// BlacklistToken adds a token to the blacklist.
func (m *InMemoryTokenBlacklistManager) BlacklistToken(ctx context.Context, token string, expiresAt time.Time, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tokens[token] = &TokenBlacklistEntry{
		Token:     token,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
		Reason:    reason,
	}
	return nil
}

// IsBlacklisted checks if a token is blacklisted.
func (m *InMemoryTokenBlacklistManager) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.tokens[token]
	if !exists {
		return false, nil
	}
	return time.Now().Before(entry.ExpiresAt), nil
}

// RemoveToken removes a token from the blacklist.
func (m *InMemoryTokenBlacklistManager) RemoveToken(ctx context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, token)
	return nil
}

// GetBlacklistedCount returns the number of blacklisted tokens.
func (m *InMemoryTokenBlacklistManager) GetBlacklistedCount(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tokens), nil
}

// GetTokenInfo returns information about a blacklisted token.
func (m *InMemoryTokenBlacklistManager) GetTokenInfo(ctx context.Context, token string) (*TokenBlacklistEntry, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.tokens[token]
	if !exists {
		return nil, false, nil
	}

	return &TokenBlacklistEntry{
		Token:     entry.Token,
		ExpiresAt: entry.ExpiresAt,
		RevokedAt: entry.RevokedAt,
		Reason:    entry.Reason,
	}, true, nil
}

// ClearAll removes all tokens from blacklist.
func (m *InMemoryTokenBlacklistManager) ClearAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens = make(map[string]*TokenBlacklistEntry)
	return nil
}

// cleanup periodically removes expired tokens.
func (m *InMemoryTokenBlacklistManager) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for token, entry := range m.tokens {
			if now.After(entry.ExpiresAt) {
				delete(m.tokens, token)
			}
		}
		m.mu.Unlock()
	}
}
