package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	auth "github.com/refynehq/refyne-backend/internal/domains/auth/utils"
)

// TestInMemoryTokenBlacklistManager tests the in-memory blacklist implementation.
func TestInMemoryTokenBlacklistManager(t *testing.T) {
	ctx := context.Background()
	manager := auth.GetInMemoryTokenBlacklistManager()

	// Clean up before test
	err := manager.ClearAll(ctx)
	require.NoError(t, err)

	t.Run("BlacklistToken_ShouldAddToken", func(t *testing.T) {
		token := "test-token-1"
		expiresAt := time.Now().Add(1 * time.Hour)
		reason := "logout"

		err := manager.BlacklistToken(ctx, token, expiresAt, reason)
		require.NoError(t, err)

		isBlacklisted, err := manager.IsBlacklisted(ctx, token)
		require.NoError(t, err)
		assert.True(t, isBlacklisted)
	})

	t.Run("IsBlacklisted_ShouldReturnFalseForUnknownToken", func(t *testing.T) {
		isBlacklisted, err := manager.IsBlacklisted(ctx, "unknown-token")
		require.NoError(t, err)
		assert.False(t, isBlacklisted)
	})

	t.Run("IsBlacklisted_ShouldReturnFalseForExpiredToken", func(t *testing.T) {
		token := "expired-token"
		expiresAt := time.Now().Add(-1 * time.Hour) // Already expired

		err := manager.BlacklistToken(ctx, token, expiresAt, "logout")
		require.NoError(t, err)

		isBlacklisted, err := manager.IsBlacklisted(ctx, token)
		require.NoError(t, err)
		assert.False(t, isBlacklisted)
	})

	t.Run("RemoveToken_ShouldRemoveFromBlacklist", func(t *testing.T) {
		token := "removable-token"
		expiresAt := time.Now().Add(1 * time.Hour)

		err := manager.BlacklistToken(ctx, token, expiresAt, "logout")
		require.NoError(t, err)

		err = manager.RemoveToken(ctx, token)
		require.NoError(t, err)

		isBlacklisted, err := manager.IsBlacklisted(ctx, token)
		require.NoError(t, err)
		assert.False(t, isBlacklisted)
	})

	t.Run("GetTokenInfo_ShouldReturnEntryDetails", func(t *testing.T) {
		token := "info-token"
		expiresAt := time.Now().Add(1 * time.Hour)
		reason := "security"

		err := manager.BlacklistToken(ctx, token, expiresAt, reason)
		require.NoError(t, err)

		entry, exists, err := manager.GetTokenInfo(ctx, token)
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, token, entry.Token)
		assert.Equal(t, reason, entry.Reason)
	})

	t.Run("GetBlacklistedCount_ShouldReturnCorrectCount", func(t *testing.T) {
		// Clear first
		err := manager.ClearAll(ctx)
		require.NoError(t, err)

		// Add 3 tokens
		for i := 1; i <= 3; i++ {
			token := "count-token-" + string(rune('0'+i))
			expiresAt := time.Now().Add(1 * time.Hour)
			err := manager.BlacklistToken(ctx, token, expiresAt, "logout")
			require.NoError(t, err)
		}

		count, err := manager.GetBlacklistedCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("ClearAll_ShouldRemoveAllTokens", func(t *testing.T) {
		// Add a token
		err := manager.BlacklistToken(ctx, "clear-token", time.Now().Add(1*time.Hour), "logout")
		require.NoError(t, err)

		err = manager.ClearAll(ctx)
		require.NoError(t, err)

		count, err := manager.GetBlacklistedCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

// TestInMemoryOTPManager tests the in-memory OTP implementation.
func TestInMemoryOTPManager(t *testing.T) {
	ctx := context.Background()
	manager := auth.GetInMemoryOTPManager()

	t.Run("GenerateOTP_ShouldReturn6DigitCode", func(t *testing.T) {
		otp, err := manager.GenerateOTP()
		require.NoError(t, err)
		assert.Len(t, otp, 6)
		// Verify it's numeric
		for _, c := range otp {
			assert.True(t, c >= '0' && c <= '9', "OTP should only contain digits")
		}
	})

	t.Run("StoreOTP_ShouldStoreAndRetrieve", func(t *testing.T) {
		email := "test@example.com"
		otp := "123456"

		err := manager.StoreOTP(ctx, email, otp)
		require.NoError(t, err)

		entry, err := manager.GetOTPInfo(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, entry)
		assert.Equal(t, otp, entry.OTP)
		assert.Equal(t, email, entry.Email)
	})

	t.Run("StoreOTP_ShouldOverwritePreviousOTP", func(t *testing.T) {
		email := "overwrite@example.com"
		otp1 := "111111"
		otp2 := "222222"

		err := manager.StoreOTP(ctx, email, otp1)
		require.NoError(t, err)

		err = manager.StoreOTP(ctx, email, otp2)
		require.NoError(t, err)

		entry, err := manager.GetOTPInfo(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, entry)
		assert.Equal(t, otp2, entry.OTP, "Should have the newer OTP")
	})

	t.Run("InvalidateOTP_ShouldRemoveOTP", func(t *testing.T) {
		email := "invalidate@example.com"
		otp := "333333"

		err := manager.StoreOTP(ctx, email, otp)
		require.NoError(t, err)

		err = manager.InvalidateOTP(ctx, email)
		require.NoError(t, err)

		entry, err := manager.GetOTPInfo(ctx, email)
		require.NoError(t, err)
		assert.Nil(t, entry)
	})

	t.Run("GetOTPInfo_ShouldReturnNilForUnknownEmail", func(t *testing.T) {
		entry, err := manager.GetOTPInfo(ctx, "unknown@example.com")
		require.NoError(t, err)
		assert.Nil(t, entry)
	})
}

// TestLegacyBlacklistManager tests the backward-compatible wrapper.
func TestLegacyBlacklistManager(t *testing.T) {
	manager := auth.GetTokenBlacklistManager()

	// Clean up
	manager.ClearAll()

	t.Run("BlacklistAndCheck", func(t *testing.T) {
		token := "legacy-token"
		expiresAt := time.Now().Add(1 * time.Hour)

		manager.BlacklistToken(token, expiresAt, "logout")

		assert.True(t, manager.IsBlacklisted(token))
		assert.False(t, manager.IsBlacklisted("other-token"))
	})

	t.Run("GetTokenInfo", func(t *testing.T) {
		token := "legacy-info-token"
		expiresAt := time.Now().Add(1 * time.Hour)

		manager.BlacklistToken(token, expiresAt, "security")

		entry, exists := manager.GetTokenInfo(token)
		assert.True(t, exists)
		assert.Equal(t, "security", entry.Reason)
	})

	t.Run("RemoveAndClear", func(t *testing.T) {
		manager.BlacklistToken("token1", time.Now().Add(1*time.Hour), "logout")
		manager.BlacklistToken("token2", time.Now().Add(1*time.Hour), "logout")

		manager.RemoveToken("token1")
		assert.False(t, manager.IsBlacklisted("token1"))
		assert.True(t, manager.IsBlacklisted("token2"))

		manager.ClearAll()
		assert.Equal(t, 0, manager.GetBlacklistedCount())
	})
}

// TestLegacyOTPManager tests the backward-compatible OTP wrapper.
func TestLegacyOTPManager(t *testing.T) {
	manager := auth.GetOTPManager()

	t.Run("GenerateAndStore", func(t *testing.T) {
		email := "legacy-otp@example.com"
		otp, err := manager.GenerateOTP()
		require.NoError(t, err)
		assert.Len(t, otp, 6)

		manager.StoreOTP(email, otp)

		entry := manager.GetOTPInfo(email)
		require.NotNil(t, entry)
		assert.Equal(t, otp, entry.OTP)
	})

	t.Run("Invalidate", func(t *testing.T) {
		email := "legacy-invalidate@example.com"
		manager.StoreOTP(email, "999999")

		manager.InvalidateOTP(email)

		entry := manager.GetOTPInfo(email)
		assert.Nil(t, entry)
	})
}
