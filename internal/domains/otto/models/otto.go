package models

import (
	"time"
)

// OttoSettings represents the configuration for the Otto AI agent
type OttoSettings struct {
	UserID             string    `json:"user_id" db:"user_id"`
	AutoReplyEnabled   bool      `json:"auto_reply_enabled" db:"auto_reply_enabled"`
	Tone               string    `json:"tone" db:"tone"`
	Guardrails         string    `json:"guardrails" db:"guardrails"`
	CustomInstructions string    `json:"custom_instructions" db:"custom_instructions"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateOttoSettingsInput represents the payload for updating settings
type UpdateOttoSettingsInput struct {
	AutoReplyEnabled   bool   `json:"auto_reply_enabled"`
	Tone               string `json:"tone" binding:"omitempty,oneof=professional casual witty sarcastic supportive"`
	Guardrails         string `json:"guardrails"`
	CustomInstructions string `json:"custom_instructions"`
}
