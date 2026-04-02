package user

import (
	"regexp"
	"strings"
	"time"
)

type User struct {
	ID           string `json:"id" db:"id"`
	Email        string `json:"email" db:"email"`
	PasswordHash string `json:"password_hash" db:"password_hash"`

	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
	Username  string `json:"username" db:"username"`

	// Account Status flags
	Status     string `json:"status" db:"status"` // e.g., "active", "inactive", "suspended"
	IsActive   bool   `json:"is_active" db:"is_active"`
	IsVerified bool   `json:"is_verified" db:"is_verified"`

	// Security fields
	LastLogin             *time.Time `json:"last_login" db:"last_login"`
	LastLoginIP           *string    `json:"last_login_ip" db:"last_login_ip"`
	LastPasswordChangedAt *time.Time `json:"last_password_changed_at,omitempty" db:"last_password_changed_at"`
	TokenVersion          int        `json:"token_version" db:"token_version"`

	// Subscription fields (Paddle integration)
	SubscriptionTier      string     `json:"subscription_tier" db:"subscription_tier"`     // starter, professional, business, enterprise
	SubscriptionStatus    string     `json:"subscription_status" db:"subscription_status"` // active, cancelled, past_due, trialing, paused, inactive
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty" db:"subscription_expires_at"`
	PaddleCustomerID      *string    `json:"paddle_customer_id,omitempty" db:"paddle_customer_id"`
	PaddleSubscriptionID  *string    `json:"paddle_subscription_id,omitempty" db:"paddle_subscription_id"`
	OnboardingCompleted   bool       `json:"onboarding_completed" db:"onboarding_completed"`

	// Timestamps for creation and updates
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // Nullable for soft delete
}

// Helper functions
func (u *User) GetStatus() string {
	return u.Status
}

func (u *User) IsActiveUser() bool {
	return u.IsActive && u.Status == "active"
}

func (u *User) IsVerifiedUser() bool {
	return u.IsVerified
}

// HasActiveSubscription checks if user has an active paid subscription
func (u *User) HasActiveSubscription() bool {
	return (u.SubscriptionStatus == "active" || u.SubscriptionStatus == "trialing") && u.SubscriptionTier == "pro"
}

// IsSubscriptionExpired checks if subscription has expired
func (u *User) IsSubscriptionExpired() bool {
	if u.SubscriptionExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.SubscriptionExpiresAt)
}

// GetSubscriptionTier returns the user's subscription tier
func (u *User) GetSubscriptionTier() string {
	return u.SubscriptionTier
}

// CanAccessFeature checks if user's tier allows access to a feature
// Currently only Pro tier is supported
func (u *User) CanAccessFeature(requiredTier string) bool {
	// All features require Pro subscription
	return u.HasActiveSubscription()
}

// Validations
func (u *User) HasValidEmail() bool {
	email := strings.TrimSpace(u.Email)
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)

	return regex.MatchString(email) && len(email) <= 254
}
