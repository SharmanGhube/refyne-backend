package models

import (
	"database/sql"
	"time"
)

// InstagramAccount represents a connected Instagram business account
type InstagramAccount struct {
	ID                string         `db:"id" json:"id"`
	UserID            string         `db:"user_id" json:"user_id"`
	InstagramUserID   string         `db:"instagram_user_id" json:"instagram_user_id"`
	Username          string         `db:"username" json:"username"`
	AccessToken       string         `db:"access_token" json:"-"` // Never expose in JSON
	RefreshToken      sql.NullString `db:"refresh_token" json:"-"`
	TokenExpiresAt    time.Time      `db:"token_expires_at" json:"token_expires_at"`
	ProfilePictureURL sql.NullString `db:"profile_picture_url" json:"profile_picture_url"`
	Biography         sql.NullString `db:"biography" json:"biography"`
	FollowersCount    int            `db:"followers_count" json:"followers_count"`

	// Account status
	ConnectedAt      time.Time      `db:"connected_at" json:"connected_at"`
	LastSyncAt       sql.NullTime   `db:"last_sync_at" json:"last_sync_at"`
	SyncStatus       string         `db:"sync_status" json:"sync_status"` // idle, syncing, or error
	SyncErrorMessage sql.NullString `db:"sync_error_message" json:"sync_error_message"`

	// Timestamps
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" json:"deleted_at"`
}

// CreateInstagramAccountInput represents input for creating a new Instagram account connection
type CreateInstagramAccountInput struct {
	UserID            string
	InstagramUserID   string
	Username          string
	AccessToken       string
	RefreshToken      string
	TokenExpiresAt    time.Time
	ProfilePictureURL string
	Biography         string
	FollowersCount    int
}

// UpdateInstagramAccountInput represents input for updating an Instagram account
type UpdateInstagramAccountInput struct {
	AccessToken       string
	RefreshToken      string
	TokenExpiresAt    time.Time
	ProfilePictureURL string
	Biography         string
	FollowersCount    int
	SyncStatus        string
	SyncErrorMessage  string
}
