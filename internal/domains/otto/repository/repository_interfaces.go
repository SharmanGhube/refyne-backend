package repository

import (
	"context"

	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
)

// OttoConversationRepository handles database operations for conversations
type OttoConversationRepository interface {
	// Create a new conversation
	CreateConversation(ctx context.Context, conversation *models.OttoConversation) error

	// Get conversation by ID
	GetConversation(ctx context.Context, conversationID string) (*models.OttoConversation, error)

	// List conversations for a user/workspace
	ListConversations(ctx context.Context, userID, workspaceID string, limit int, offset int) ([]*models.OttoConversation, error)

	// Update conversation
	UpdateConversation(ctx context.Context, conversation *models.OttoConversation) error

	// Archive conversation
	ArchiveConversation(ctx context.Context, conversationID string) error

	// Delete conversation (soft delete)
	DeleteConversation(ctx context.Context, conversationID string) error

	// Toggle bookmark
	SetBookmarked(ctx context.Context, conversationID string, isBookmarked bool) error
}

// OttoMessageRepository handles database operations for messages
type OttoMessageRepository interface {
	// Create a new message
	CreateMessage(ctx context.Context, message *models.OttoMessage) error

	// Get message by ID
	GetMessage(ctx context.Context, messageID string) (*models.OttoMessage, error)

	// Get all messages for a conversation
	GetConversationMessages(ctx context.Context, conversationID string, limit int, offset int) ([]*models.OttoMessage, error)

	// Update message (e.g., add feedback)
	UpdateMessage(ctx context.Context, message *models.OttoMessage) error

	// Add feedback to a message
	AddFeedback(ctx context.Context, messageID string, isLiked bool, notes string) error

	// Get conversation message count
	GetMessageCount(ctx context.Context, conversationID string) (int, error)

	// Delete message (soft delete)
	DeleteMessage(ctx context.Context, messageID string) error
}
