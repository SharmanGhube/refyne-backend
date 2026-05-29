package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// --- DefaultPasswordPolicy Tests ---

func TestDefaultPasswordPolicy(t *testing.T) {
	policy := DefaultPasswordPolicy()

	assert.Equal(t, 8, policy.MinLength)
	assert.True(t, policy.RequireUpper)
	assert.True(t, policy.RequireLower)
	assert.True(t, policy.RequireDigit)
	assert.True(t, policy.RequireSpecial)
}

// --- PasswordPolicy.Validate Tests ---

func TestPasswordPolicy_Validate(t *testing.T) {
	policy := DefaultPasswordPolicy()

	tests := []struct {
		name      string
		password  string
		wantError bool
		errMsg    string
	}{
		{
			name:      "valid password",
			password:  "MyP@ssw0rd!",
			wantError: false,
		},
		{
			name:      "too short",
			password:  "Ab1!xyz",
			wantError: true,
			errMsg:    "at least 8 characters",
		},
		{
			name:      "no uppercase",
			password:  "myp@ssw0rd!",
			wantError: true,
			errMsg:    "uppercase",
		},
		{
			name:      "no lowercase",
			password:  "MYP@SSW0RD!",
			wantError: true,
			errMsg:    "lowercase",
		},
		{
			name:      "no digit",
			password:  "MyP@ssword!",
			wantError: true,
			errMsg:    "digit",
		},
		{
			name:      "no special character",
			password:  "MyPassw0rd1",
			wantError: true,
			errMsg:    "special",
		},
		{
			name:      "minimum valid password",
			password:  "Aa1!xxxx",
			wantError: false,
		},
		{
			name:      "empty password",
			password:  "",
			wantError: true,
			errMsg:    "at least 8 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.password)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPasswordPolicy_Validate_RelaxedPolicy(t *testing.T) {
	// Test with all requirements disabled
	policy := PasswordPolicy{
		MinLength:      4,
		RequireUpper:   false,
		RequireLower:   false,
		RequireDigit:   false,
		RequireSpecial: false,
	}

	err := policy.Validate("abcd")
	assert.NoError(t, err, "should accept simple password with relaxed policy")

	err = policy.Validate("abc")
	assert.Error(t, err, "should still enforce min length")
}

// --- GenerateHash Tests ---

func TestGenerateHash_ValidCost(t *testing.T) {
	hash, err := GenerateHash("testpassword", bcrypt.DefaultCost)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "testpassword", hash)
	// bcrypt hashes start with $2a$ or $2b$
	assert.Regexp(t, `^\$2[ab]\$`, hash)
}

func TestGenerateHash_MinCost(t *testing.T) {
	hash, err := GenerateHash("testpassword", bcrypt.MinCost)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

func TestGenerateHash_InvalidCost_TooLow(t *testing.T) {
	hash, err := GenerateHash("testpassword", bcrypt.MinCost-1)

	assert.Error(t, err)
	assert.Empty(t, hash)
	assert.Contains(t, err.Error(), "invalid cost")
}

func TestGenerateHash_InvalidCost_TooHigh(t *testing.T) {
	hash, err := GenerateHash("testpassword", bcrypt.MaxCost+1)

	assert.Error(t, err)
	assert.Empty(t, hash)
	assert.Contains(t, err.Error(), "invalid cost")
}

// --- CheckHash Tests ---

func TestCheckHash_CorrectPassword(t *testing.T) {
	password := "MySecureP@ss1"
	hash, err := GenerateHash(password, bcrypt.MinCost) // Use MinCost for fast tests
	require.NoError(t, err)

	match, err := CheckHash(password, hash)

	assert.NoError(t, err)
	assert.True(t, match)
}

func TestCheckHash_WrongPassword(t *testing.T) {
	hash, err := GenerateHash("correctpassword", bcrypt.MinCost)
	require.NoError(t, err)

	match, err := CheckHash("wrongpassword", hash)

	assert.Error(t, err)
	assert.False(t, match)
	assert.Contains(t, err.Error(), "does not match")
}

func TestCheckHash_EmptyPassword(t *testing.T) {
	hash, err := GenerateHash("somepassword", bcrypt.MinCost)
	require.NoError(t, err)

	match, err := CheckHash("", hash)

	assert.Error(t, err)
	assert.False(t, match)
}
