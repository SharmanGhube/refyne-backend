package user

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- GetStatus ---

func TestGetStatus_ReturnsCorrectStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{"active status", "active", "active"},
		{"inactive status", "inactive", "inactive"},
		{"suspended status", "suspended", "suspended"},
		{"empty status", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Status: tt.status}
			assert.Equal(t, tt.expected, user.GetStatus())
		})
	}
}

// --- IsActiveUser ---

func TestIsActiveUser(t *testing.T) {
	tests := []struct {
		name     string
		isActive bool
		status   string
		expected bool
	}{
		{"active flag and active status", true, "active", true},
		{"active flag but inactive status", true, "inactive", false},
		{"active flag but suspended status", true, "suspended", false},
		{"inactive flag but active status", false, "active", false},
		{"inactive flag and inactive status", false, "inactive", false},
		{"active flag and empty status", true, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{IsActive: tt.isActive, Status: tt.status}
			assert.Equal(t, tt.expected, user.IsActiveUser())
		})
	}
}

// --- IsVerifiedUser ---

func TestIsVerifiedUser(t *testing.T) {
	tests := []struct {
		name       string
		isVerified bool
		expected   bool
	}{
		{"verified user", true, true},
		{"unverified user", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{IsVerified: tt.isVerified}
			assert.Equal(t, tt.expected, user.IsVerifiedUser())
		})
	}
}

// --- HasActiveSubscription ---

func TestHasActiveSubscription(t *testing.T) {
	tests := []struct {
		name   string
		tier   string
		status string
		want   bool
	}{
		{"active pro", "pro", "active", true},
		{"trialing pro", "pro", "trialing", true},
		{"cancelled pro", "pro", "cancelled", false},
		{"past_due pro", "pro", "past_due", false},
		{"paused pro", "pro", "paused", false},
		{"active starter (non-pro)", "starter", "active", false},
		{"active empty tier", "", "active", false},
		{"empty status pro", "pro", "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				SubscriptionTier:   tt.tier,
				SubscriptionStatus: tt.status,
			}
			assert.Equal(t, tt.want, user.HasActiveSubscription())
		})
	}
}

// --- IsSubscriptionExpired ---

func TestIsSubscriptionExpired(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	futureTime := now.Add(24 * time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		want      bool
	}{
		{"nil expiry — not expired", nil, false},
		{"future expiry — not expired", &futureTime, false},
		{"past expiry — expired", &pastTime, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.want, user.IsSubscriptionExpired())
		})
	}
}

// --- GetSubscriptionTier ---

func TestGetSubscriptionTier(t *testing.T) {
	tests := []struct {
		name string
		tier string
		want string
	}{
		{"pro tier", "pro", "pro"},
		{"starter tier", "starter", "starter"},
		{"empty tier", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{SubscriptionTier: tt.tier}
			assert.Equal(t, tt.want, user.GetSubscriptionTier())
		})
	}
}

// --- CanAccessFeature ---

func TestCanAccessFeature(t *testing.T) {
	tests := []struct {
		name         string
		tier         string
		status       string
		requiredTier string
		want         bool
	}{
		// Current behavior: CanAccessFeature delegates to HasActiveSubscription,
		// ignoring the requiredTier parameter entirely.
		{"active pro requesting pro", "pro", "active", "pro", true},
		{"active pro requesting starter", "pro", "active", "starter", true},
		{"trialing pro requesting pro", "pro", "trialing", "pro", true},
		{"active starter requesting starter", "starter", "active", "starter", false},
		{"cancelled pro requesting pro", "pro", "cancelled", "pro", false},
		{"no subscription requesting any", "", "", "pro", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				SubscriptionTier:   tt.tier,
				SubscriptionStatus: tt.status,
			}
			assert.Equal(t, tt.want, user.CanAccessFeature(tt.requiredTier))
		})
	}
}

// --- HasValidEmail ---

func TestHasValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		// Valid emails
		{"simple valid email", "user@example.com", true},
		{"email with plus alias", "user+tag@example.com", true},
		{"email with subdomain", "user@mail.example.com", true},
		{"email with dots in local", "first.last@example.com", true},
		{"email with numbers", "user123@example123.com", true},
		{"email with hyphens in domain", "user@my-domain.com", true},

		// Invalid emails
		{"empty email", "", false},
		{"missing @", "userexample.com", false},
		{"missing domain", "user@", false},
		{"missing local part", "@example.com", false},
		{"double @", "user@@example.com", false},
		{"missing TLD", "user@example", false},
		{"spaces in email", "us er@example.com", false},
		{"single char TLD", "user@example.c", false},

		// Length boundary
		{"at max length (254 chars)", generateEmail(254), true},
		{"exceeds max length (255 chars)", generateEmail(255), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Email: tt.email}
			assert.Equal(t, tt.want, user.HasValidEmail())
		})
	}
}

// generateEmail creates an email of exactly the specified total length.
// Format: <local>@example.com (12 chars for "@example.com")
func generateEmail(length int) string {
	suffix := "@example.com" // 12 chars
	if length <= len(suffix) {
		return ""
	}
	localLen := length - len(suffix)
	local := make([]byte, localLen)
	for i := range local {
		local[i] = 'a'
	}
	return string(local) + suffix
}
