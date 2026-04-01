package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	serviceErrors "github.com/refynehq/refyne-backend/internal/domains/auth/services/errors"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// InMemoryOTPManager is the in-memory implementation.
// For production with multiple instances, use RedisOTPManager.
type InMemoryOTPManager struct {
	otps map[string]*OTPEntry
	mu   sync.RWMutex
}

var (
	inMemoryOTPManager     *InMemoryOTPManager
	inMemoryOTPManagerOnce sync.Once
)

// GetInMemoryOTPManager returns the singleton in-memory OTP manager.
func GetInMemoryOTPManager() *InMemoryOTPManager {
	inMemoryOTPManagerOnce.Do(func() {
		inMemoryOTPManager = &InMemoryOTPManager{
			otps: make(map[string]*OTPEntry),
		}
		go inMemoryOTPManager.cleanup()
	})
	return inMemoryOTPManager
}

// GenerateOTP creates a 6-digit numeric OTP.
func (m *InMemoryOTPManager) GenerateOTP() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	num := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	otp := fmt.Sprintf("%06d", num%1000000)
	return otp, nil
}

// StoreOTP stores an OTP for an email with 15-minute expiry.
func (m *InMemoryOTPManager) StoreOTP(ctx context.Context, email, otp string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Invalidate any existing OTP for this email
	delete(m.otps, email)

	// Store new OTP
	m.otps[email] = &OTPEntry{
		OTP:       otp,
		Email:     email,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
	return nil
}

// ValidateOTP validates an OTP for an email.
func (m *InMemoryOTPManager) ValidateOTP(c *gin.Context, email, otp string) *errors.AppError {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.otps[email]
	if !exists {
		return serviceErrors.NewOTPNotFoundError(c, email)
	}

	// Check if OTP has expired
	if time.Now().After(entry.ExpiresAt) {
		go func() {
			m.mu.Lock()
			delete(m.otps, email)
			m.mu.Unlock()
		}()
		return serviceErrors.NewOTPExpiredError(c, email)
	}

	// Validate OTP
	if entry.OTP != otp {
		return serviceErrors.NewInvalidOTPError(c, "Invalid OTP")
	}

	return nil
}

// InvalidateOTP removes an OTP from memory.
func (m *InMemoryOTPManager) InvalidateOTP(ctx context.Context, email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.otps, email)
	return nil
}

// GetOTPInfo returns OTP info (for debugging/testing).
func (m *InMemoryOTPManager) GetOTPInfo(ctx context.Context, email string) (*OTPEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if entry, exists := m.otps[email]; exists {
		return &OTPEntry{
			OTP:       entry.OTP,
			Email:     entry.Email,
			ExpiresAt: entry.ExpiresAt,
			CreatedAt: entry.CreatedAt,
		}, nil
	}
	return nil, nil
}

// cleanup periodically removes expired OTPs.
func (m *InMemoryOTPManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for email, entry := range m.otps {
			if now.After(entry.ExpiresAt) {
				delete(m.otps, email)
			}
		}
		m.mu.Unlock()
	}
}
