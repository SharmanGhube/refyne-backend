package user

// User model represents a user in the system.
// Took me 3 days to come up with this fucking model for no absolute reason
type User struct {
	ID           string `json:"id" db:"id"`
	Email        string `json:"email" db:"email"`
	Username     string `json:"username" db:"username"`
	PasswordHash string `json:"-" db:"password_hash"` // Password should not be returned in responses

	// Account Status flags
	Status     string `json:"status" db:"status"` // (Pending, Active, Banned, Deleted)
	IsActive   bool   `json:"is_active" db:"is_active"`
	IsVerified bool   `json:"is_verified" db:"is_verified"`

	TimeZone string `json:"timezone" db:"timezone"` // User's timezone for scheduling and notifications

	// Security fields
	LastLoginAt string `json:"last_login_at" db:"last_login_at"` // Timestamp of the last login
	LastLoginIP string `json:"last_login_ip" db:"last_login_ip"` // IP address of the last login

	// Timestamps
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
	DeletedAt string `json:"deleted_at,omitempty" db:"deleted_at"` // Nullable, for soft deletes
}

// Helper functions for struct
func (u *User) IsActiveUser() bool {
	return u.IsActive && u.IsVerified
}
