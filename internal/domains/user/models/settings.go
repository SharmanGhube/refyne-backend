package user

import "time"

type UserSettings struct {
	ID     string `json:"id" db:"id"`
	UserID string `json:"user_id" db:"user_id"`

	// Localization settings
	Language string `json:"language" db:"language"`
	TimeZone string `json:"time_zone" db:"time_zone"`

	// Notification preferences
	EmailNotifications bool `json:"email_notifications" db:"email_notifications"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
