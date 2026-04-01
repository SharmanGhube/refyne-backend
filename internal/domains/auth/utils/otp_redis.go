package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	serviceErrors "github.com/refynehq/refyne-backend/internal/domains/auth/services/errors"
	errors "github.com/refynehq/refyne-backend/pkg/error"
)

// OTPManager defines the interface for OTP operations.
type OTPManager interface {
	GenerateOTP() (string, error)
	StoreOTP(ctx context.Context, email, otp string) error
	ValidateOTP(c *gin.Context, email, otp string) *errors.AppError
	InvalidateOTP(ctx context.Context, email string) error
	GetOTPInfo(ctx context.Context, email string) (*OTPEntry, error)
}

// OTPEntry represents an OTP entry.
type OTPEntry struct {
	OTP       string    `json:"otp"`
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

const (
	otpKeyPrefix  = "otp:"
	otpDefaultTTL = 15 * time.Minute
)

// RedisOTPManager implements OTPManager using Redis.
type RedisOTPManager struct {
	client *redis.Client
}

// NewRedisOTPManager creates a new Redis-backed OTP manager.
func NewRedisOTPManager(client *redis.Client) *RedisOTPManager {
	return &RedisOTPManager{
		client: client,
	}
}

// GenerateOTP creates a 6-digit numeric OTP.
func (m *RedisOTPManager) GenerateOTP() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	num := int(bytes[0])<<16 | int(bytes[1])<<8 | int(bytes[2])
	otp := fmt.Sprintf("%06d", num%1000000)
	return otp, nil
}

// StoreOTP stores an OTP for an email with 15-minute expiry.
func (m *RedisOTPManager) StoreOTP(ctx context.Context, email, otp string) error {
	entry := OTPEntry{
		OTP:       otp,
		Email:     email,
		ExpiresAt: time.Now().Add(otpDefaultTTL),
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal OTP entry: %w", err)
	}

	key := otpKeyPrefix + email
	if err := m.client.Set(ctx, key, data, otpDefaultTTL).Err(); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	return nil
}

// ValidateOTP validates an OTP for an email.
func (m *RedisOTPManager) ValidateOTP(c *gin.Context, email, otp string) *errors.AppError {
	ctx := c.Request.Context()
	key := otpKeyPrefix + email

	data, err := m.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return serviceErrors.NewOTPNotFoundError(c, email)
	}
	if err != nil {
		return errors.NewAppError(
			c,
			"OTP_VALIDATION_ERROR",
			"Failed to validate OTP",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"auth",
		)
	}

	var entry OTPEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return errors.NewAppError(
			c,
			"OTP_PARSE_ERROR",
			"Failed to parse OTP data",
			errors.ErrorTypeInternal,
			errors.SeverityHigh,
			"auth",
		)
	}

	// Check if OTP has expired (belt and suspenders - Redis TTL should handle this)
	if time.Now().After(entry.ExpiresAt) {
		_ = m.InvalidateOTP(ctx, email)
		return serviceErrors.NewOTPExpiredError(c, email)
	}

	// Validate OTP
	if entry.OTP != otp {
		return serviceErrors.NewInvalidOTPError(c, "Invalid OTP")
	}

	return nil
}

// InvalidateOTP removes an OTP from storage.
func (m *RedisOTPManager) InvalidateOTP(ctx context.Context, email string) error {
	key := otpKeyPrefix + email
	if err := m.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate OTP: %w", err)
	}
	return nil
}

// GetOTPInfo returns OTP info (for debugging/testing).
func (m *RedisOTPManager) GetOTPInfo(ctx context.Context, email string) (*OTPEntry, error) {
	key := otpKeyPrefix + email

	data, err := m.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OTP info: %w", err)
	}

	var entry OTPEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("failed to parse OTP entry: %w", err)
	}

	return &entry, nil
}
