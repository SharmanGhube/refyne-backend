package user

import (
	"time"

	"github.com/google/uuid"
)

type UserSettings struct {
	ID     string `json:"id" db:"id"`
	UserID string `json:"user_id" db:"user_id"`

	// Localization settings
	Language string `json:"language" db:"language"`
	Timezone string `json:"timezone" db:"timezone"`

	// Notification settings
	EmailNotifications bool `json:"email_notifications" db:"email_notifications"`

	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
}

// NewDefaultUserSettings creates a new UserSettings with default values
func NewDefaultUserSettings(userID string) *UserSettings {
	now := time.Now()
	return &UserSettings{
		ID:                 uuid.New().String(),
		UserID:             userID,
		Language:           "en",
		Timezone:           "UTC",
		EmailNotifications: true,
		CreatedAt:          &now,
		UpdatedAt:          &now,
	}
}
