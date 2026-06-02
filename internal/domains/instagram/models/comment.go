package models

import (
	"time"
)

// InstagramComment represents an Instagram comment on a media post
type InstagramComment struct {
	ID               string    `db:"id" json:"id"`
	InstagramMediaID string    `db:"instagram_media_id" json:"instagram_media_id"`
	AccountID        string    `db:"account_id" json:"account_id"`
	Username         string    `db:"username" json:"username"`
	Text             string    `db:"text" json:"text"`
	IsHidden         bool      `db:"is_hidden" json:"is_hidden"`
	IsFlagged        bool      `db:"is_flagged" json:"is_flagged"`
	ModerationReason string    `db:"moderation_reason" json:"moderation_reason,omitempty"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}
