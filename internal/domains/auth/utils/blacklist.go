package auth

import (
	"sync"
	"time"
)

// TokenBlacklistEntry represents a blacklisted token with expiry
type TokenBlacklistEntry struct {
	Token     string
	ExpiresAt time.Time
	RevokedAt time.Time
	Reason    string // "logout", "password_reset", "security"
}

// TokenBlacklistManager manages revoked/blacklisted JWT tokens
// This is an in-memory implementation - for production with multiple instances,
// consider using Redis or a database-backed solution
type TokenBlacklistManager struct {
	tokens map[string]*TokenBlacklistEntry
	mu     sync.RWMutex
}

var (
	blacklistManager     *TokenBlacklistManager
	blacklistManagerOnce sync.Once
)

// GetTokenBlacklistManager returns the singleton token blacklist manager
func GetTokenBlacklistManager() *TokenBlacklistManager {
	blacklistManagerOnce.Do(func() {
		blacklistManager = &TokenBlacklistManager{
			tokens: make(map[string]*TokenBlacklistEntry),
		}
		// Start cleanup goroutine
		go blacklistManager.cleanup()
	})
	return blacklistManager
}

// BlacklistToken adds a token to the blacklist
func (tbm *TokenBlacklistManager) BlacklistToken(token string, expiresAt time.Time, reason string) {
	tbm.mu.Lock()
	defer tbm.mu.Unlock()

	tbm.tokens[token] = &TokenBlacklistEntry{
		Token:     token,
		ExpiresAt: expiresAt,
		RevokedAt: time.Now(),
		Reason:    reason,
	}
}

// IsBlacklisted checks if a token is blacklisted
func (tbm *TokenBlacklistManager) IsBlacklisted(token string) bool {
	tbm.mu.RLock()
	defer tbm.mu.RUnlock()

	entry, exists := tbm.tokens[token]
	if !exists {
		return false
	}

	// Token is blacklisted and not yet expired
	return time.Now().Before(entry.ExpiresAt)
}

// RemoveToken removes a token from the blacklist (admin function)
func (tbm *TokenBlacklistManager) RemoveToken(token string) {
	tbm.mu.Lock()
	defer tbm.mu.Unlock()

	delete(tbm.tokens, token)
}

// GetBlacklistedCount returns the number of blacklisted tokens
func (tbm *TokenBlacklistManager) GetBlacklistedCount() int {
	tbm.mu.RLock()
	defer tbm.mu.RUnlock()

	return len(tbm.tokens)
}

// cleanup periodically removes expired tokens from the blacklist
func (tbm *TokenBlacklistManager) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		tbm.mu.Lock()
		now := time.Now()

		// Remove expired tokens
		for token, entry := range tbm.tokens {
			if now.After(entry.ExpiresAt) {
				delete(tbm.tokens, token)
			}
		}

		tbm.mu.Unlock()
	}
}

// ClearAll removes all tokens from blacklist (for testing/admin purposes)
func (tbm *TokenBlacklistManager) ClearAll() {
	tbm.mu.Lock()
	defer tbm.mu.Unlock()

	tbm.tokens = make(map[string]*TokenBlacklistEntry)
}

// GetTokenInfo returns information about a blacklisted token
func (tbm *TokenBlacklistManager) GetTokenInfo(token string) (*TokenBlacklistEntry, bool) {
	tbm.mu.RLock()
	defer tbm.mu.RUnlock()

	entry, exists := tbm.tokens[token]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	return &TokenBlacklistEntry{
		Token:     entry.Token,
		ExpiresAt: entry.ExpiresAt,
		RevokedAt: entry.RevokedAt,
		Reason:    entry.Reason,
	}, true
}
