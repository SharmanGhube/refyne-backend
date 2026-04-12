package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// OttoConversation represents an AI assistant conversation thread
type OttoConversation struct {
	ID          string         `db:"id" json:"id"`
	UserID      string         `db:"user_id" json:"user_id"`
	WorkspaceID string         `db:"workspace_id" json:"workspace_id"`
	Title       string         `db:"title" json:"title"`
	Description sql.NullString `db:"description" json:"description"`
	Context     string         `db:"context" json:"context"` // JSON context (accounts, metrics, documents)

	// Conversation state
	Status    string    `db:"status" json:"status"`       // active, archived, deleted
	IsBookmarked bool    `db:"is_bookmarked" json:"is_bookmarked"`

	// Metadata
	MessageCount int       `db:"message_count" json:"message_count"`
	LastMessageAt time.Time `db:"last_message_at" json:"last_message_at"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// OttoMessage represents a single message in a conversation
type OttoMessage struct {
	ID             string    `db:"id" json:"id"`
	ConversationID string    `db:"conversation_id" json:"conversation_id"`
	UserID         string    `db:"user_id" json:"user_id"`

	// Message content
	Role    string `db:"role" json:"role"`        // "user" or "assistant"
	Content string `db:"content" json:"content"`  // Message text

	// Metadata
	TokensUsed int           `db:"tokens_used" json:"tokens_used"`
	ModelUsed  string        `db:"model_used" json:"model_used"` // Which AI model responded
	Metadata   sql.NullString `db:"metadata" json:"metadata"`     // Additional context (JSON)

	// Feedback
	IsLiked    sql.NullBool  `db:"is_liked" json:"is_liked"`
	FeedbackNotes sql.NullString `db:"feedback_notes" json:"feedback_notes"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CreateOttoConversationInput is the request to create a new conversation
type CreateOttoConversationInput struct {
	Title       string `json:"title" binding:"required,max=255"`
	Description string `json:"description" binding:"max=1000"`
	Context     string `json:"context" binding:"required"` // JSON context
}

// CreateOttoMessageInput is the request to send a message
type CreateOttoMessageInput struct {
	ConversationID string `json:"conversation_id" binding:"required"`
	Content        string `json:"content" binding:"required,max=5000"`
}

// UpdateOttoConversationInput is the request to update a conversation
type UpdateOttoConversationInput struct {
	Title       string `json:"title" binding:"max=255"`
	Description string `json:"description" binding:"max=1000"`
	IsBookmarked bool   `json:"is_bookmarked"`
}

// OttoConversationResponse is the API response for a conversation
type OttoConversationResponse struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description,omitempty"`
	Status        string    `json:"status"`
	IsBookmarked  bool      `json:"is_bookmarked"`
	MessageCount  int       `json:"message_count"`
	LastMessageAt time.Time `json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// OttoMessageResponse is the API response for a message
type OttoMessageResponse struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	TokensUsed     int       `json:"tokens_used"`
	ModelUsed      string    `json:"model_used"`
	IsLiked        *bool     `json:"is_liked,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// ConversationContext represents the context data passed with a conversation
type ConversationContext struct {
	AccountID     string `json:"account_id,omitempty"`     // Instagram account
	PlatformType  string `json:"platform_type,omitempty"`  // instagram, tiktok, etc
	MetricsScope  string `json:"metrics_scope,omitempty"`  // last_7_days, last_30_days, all
	IncludeMedia  bool   `json:"include_media,omitempty"`  // Include recent media in context
	IncludeInsights bool  `json:"include_insights"`        // Include analytics insights
	RelatedDocuments []string `json:"related_documents,omitempty"` // Document IDs for context
}

// NewOttoConversation creates a new conversation
func NewOttoConversation(userID, workspaceID, title, description, contextJSON string) *OttoConversation {
	now := time.Now()
	return &OttoConversation{
		ID:          uuid.New().String(),
		UserID:      userID,
		WorkspaceID: workspaceID,
		Title:       title,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
		Context:   contextJSON,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewOttoMessage creates a new message
func NewOttoMessage(conversationID, userID, role, content string) *OttoMessage {
	return &OttoMessage{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		UserID:         userID,
		Role:           role,
		Content:        content,
		CreatedAt:      time.Now(),
	}
}
