package auth

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// globalOTPManager holds the active OTP manager (either Redis or in-memory).
var (
	globalOTPManager     OTPManager
	globalOTPManagerOnce sync.Once
	globalOTPManagerMu   sync.RWMutex
)

// LegacyOTPManager provides backward-compatible API.
// It wraps the new OTPManager interface.
type LegacyOTPManager struct {
	manager OTPManager
}

// InitOTPManager initializes the global OTP manager with Redis.
// Call this during application startup with a Redis client.
// If redisClient is nil, falls back to in-memory implementation.
func InitOTPManager(redisClient *redis.Client) {
	globalOTPManagerMu.Lock()
	defer globalOTPManagerMu.Unlock()

	if redisClient != nil {
		globalOTPManager = NewRedisOTPManager(redisClient)
	} else {
		globalOTPManager = GetInMemoryOTPManager()
	}
}

// GetOTPManager returns the backward-compatible OTP manager.
// This maintains the existing API signature for existing code.
func GetOTPManager() *LegacyOTPManager {
	globalOTPManagerOnce.Do(func() {
		globalOTPManagerMu.Lock()
		defer globalOTPManagerMu.Unlock()
		if globalOTPManager == nil {
			// Default to in-memory if not initialized
			globalOTPManager = GetInMemoryOTPManager()
		}
	})

	globalOTPManagerMu.RLock()
	defer globalOTPManagerMu.RUnlock()

	return &LegacyOTPManager{
		manager: globalOTPManager,
	}
}

// GenerateOTP creates a 6-digit numeric OTP.
func (l *LegacyOTPManager) GenerateOTP() (string, error) {
	return l.manager.GenerateOTP()
}

// StoreOTP stores an OTP for an email (backward-compatible API).
func (l *LegacyOTPManager) StoreOTP(email, otp string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Ignore error for backward compatibility
	_ = l.manager.StoreOTP(ctx, email, otp)
}

// ValidateOTP validates an OTP for an email (backward-compatible API).
func (l *LegacyOTPManager) ValidateOTP(c *gin.Context, email, otp string) *errors.AppError {
	return l.manager.ValidateOTP(c, email, otp)
}

// InvalidateOTP removes an OTP from storage.
func (l *LegacyOTPManager) InvalidateOTP(email string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = l.manager.InvalidateOTP(ctx, email)
}

// GetOTPInfo returns OTP info (for debugging/testing).
func (l *LegacyOTPManager) GetOTPInfo(email string) *OTPEntry {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	entry, err := l.manager.GetOTPInfo(ctx, email)
	if err != nil {
		return nil
	}
	return entry
}

// GetUnderlyingManager returns the underlying OTPManager.
// Use this for new code that wants the full interface with context and error handling.
func (l *LegacyOTPManager) GetUnderlyingManager() OTPManager {
	return l.manager
}
