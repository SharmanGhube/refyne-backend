package testutil

import (
	"time"

	"github.com/google/uuid"
	user "github.com/refynehq/refyne-backend/internal/domains/user/models"
)

// UserFixture creates a test user with sensible defaults.
// Options can be passed to customize the user.
func UserFixture(opts ...UserOption) *user.User {
	now := time.Now()
	u := &user.User{
		ID:                  uuid.New().String(),
		Email:               RandomEmail(),
		PasswordHash:        "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X4.XkAGPlFXnWJKHi", // "password123"
		FirstName:           "Test",
		LastName:            "User",
		Username:            RandomUsername(),
		Status:              "active",
		IsActive:            true,
		IsVerified:          true,
		TokenVersion:        0,
		SubscriptionTier:    "starter",
		SubscriptionStatus:  "active",
		OnboardingCompleted: false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	for _, opt := range opts {
		opt(u)
	}

	return u
}

// UserOption is a function that modifies a user fixture.
type UserOption func(*user.User)

// WithEmail sets the user's email.
func WithEmail(email string) UserOption {
	return func(u *user.User) {
		u.Email = email
	}
}

// WithUsername sets the user's username.
func WithUsername(username string) UserOption {
	return func(u *user.User) {
		u.Username = username
	}
}

// WithPasswordHash sets the user's password hash.
func WithPasswordHash(hash string) UserOption {
	return func(u *user.User) {
		u.PasswordHash = hash
	}
}

// WithUnverified sets the user as unverified.
func WithUnverified() UserOption {
	return func(u *user.User) {
		u.IsVerified = false
	}
}

// WithInactive sets the user as inactive.
func WithInactive() UserOption {
	return func(u *user.User) {
		u.IsActive = false
		u.Status = "inactive"
	}
}

// WithSubscriptionTier sets the user's subscription tier.
func WithSubscriptionTier(tier string) UserOption {
	return func(u *user.User) {
		u.SubscriptionTier = tier
	}
}

// WithSubscriptionStatus sets the user's subscription status.
func WithSubscriptionStatus(status string) UserOption {
	return func(u *user.User) {
		u.SubscriptionStatus = status
	}
}

// WithOnboardingCompleted sets the onboarding as completed.
func WithOnboardingCompleted() UserOption {
	return func(u *user.User) {
		u.OnboardingCompleted = true
	}
}

// WithTokenVersion sets the user's token version.
func WithTokenVersion(version int) UserOption {
	return func(u *user.User) {
		u.TokenVersion = version
	}
}

// RegistrationRequest creates a test registration request.
type RegistrationRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// NewRegistrationRequest creates a valid registration request.
func NewRegistrationRequest() RegistrationRequest {
	return RegistrationRequest{
		Email:     RandomEmail(),
		Password:  "SecurePass123!",
		FirstName: "Test",
		LastName:  "User",
		Username:  RandomUsername(),
	}
}

// LoginRequest creates a test login request.
type LoginRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

// NewLoginRequest creates a valid login request.
func NewLoginRequest(email, otp string) LoginRequest {
	return LoginRequest{
		Email: email,
		OTP:   otp,
	}
}
