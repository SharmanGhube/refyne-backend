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
	LastLogin   *time.Time `json:"last_login" db:"last_login"`
	LastLoginIP *string    `json:"last_login_ip" db:"last_login_ip"`

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

// Validations
func (u *User) HasValidEmail() bool {
	email := strings.TrimSpace(u.Email)
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)

	return regex.MatchString(email) && len(email) <= 254
}
