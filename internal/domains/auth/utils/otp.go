package auth

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	serviceErrors "github.com/refynehq/refyne-backend/internal/domains/auth/services/errors"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// OTPEntry represents an OTP entry in memory
type OTPEntry struct {
	OTP       string
	Email     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// OTPManager handles OTP storage and validation in memory
type OTPManager struct {
	otps map[string]*OTPEntry // key: email, value: OTPEntry
	mu   sync.RWMutex
}

// Global OTP manager instance
var otpManager *OTPManager
var otpOnce sync.Once

// GetOTPManager returns the singleton OTP manager instance
func GetOTPManager() *OTPManager {
	otpOnce.Do(func() {
		otpManager = &OTPManager{
			otps: make(map[string]*OTPEntry),
		}
		// Start cleanup goroutine
		go otpManager.cleanup()
	})
	return otpManager
}

// GenerateOTP creates a 6-digit numeric OTP
func (om *OTPManager) GenerateOTP() (string, error) {
	// Generate 6-digit OTP
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Convert to 6-digit number
	num := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	otp := fmt.Sprintf("%06d", num%1000000)
	return otp, nil
}

// StoreOTP stores an OTP for an email with 15-minute expiry
func (om *OTPManager) StoreOTP(email, otp string) {
	om.mu.Lock()
	defer om.mu.Unlock()

	// Invalidate any existing OTP for this email
	delete(om.otps, email)

	// Store new OTP
	om.otps[email] = &OTPEntry{
		OTP:       otp,
		Email:     email,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		CreatedAt: time.Now(),
	}
}

// ValidateOTP validates an OTP for an email
func (om *OTPManager) ValidateOTP(c *gin.Context, email, otp string) *errors.AppError {
	om.mu.RLock()
	defer om.mu.RUnlock()

	entry, exists := om.otps[email]
	if !exists {
		return serviceErrors.NewOTPNotFoundError(c, email)
	}

	// Check if OTP has expired
	if time.Now().After(entry.ExpiresAt) {
		// Clean up expired OTP
		go func() {
			om.mu.Lock()
			delete(om.otps, email)
			om.mu.Unlock()
		}()

		return serviceErrors.NewOTPExpiredError(c, email)
	}

	// Validate OTP
	if entry.OTP != otp {
		return serviceErrors.NewInvalidOTPError(c, "Invalid OTP")
	}

	return nil
}

// InvalidateOTP removes an OTP from memory after successful validation
func (om *OTPManager) InvalidateOTP(email string) {
	om.mu.Lock()
	defer om.mu.Unlock()
	delete(om.otps, email)
}

// cleanup periodically removes expired OTPs from memory
func (om *OTPManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Run cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		om.mu.Lock()
		now := time.Now()
		for email, entry := range om.otps {
			if now.After(entry.ExpiresAt) {
				delete(om.otps, email)
			}
		}
		om.mu.Unlock()
	}
}

// GetOTPInfo returns OTP info for debugging (non-production use)
func (om *OTPManager) GetOTPInfo(email string) *OTPEntry {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if entry, exists := om.otps[email]; exists {
		return &OTPEntry{
			OTP:       entry.OTP,
			Email:     entry.Email,
			ExpiresAt: entry.ExpiresAt,
			CreatedAt: entry.CreatedAt,
		}
	}
	return nil
}
