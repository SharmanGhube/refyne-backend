package services

import (
	"context"
	"encoding/json"

	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
	"github.com/refynehq/refyne-backend/internal/domains/otto/repository"
)

// OttoAssistantService provides AI assistance functionality
type OttoAssistantService interface {
	// ProcessMessage handles a user message and generates an AI response
	// Returns the AI response message and any error
	ProcessMessage(ctx context.Context, conversationID, userMessage string) (*models.OttoMessage, error)

	// GetConversationHistory retrieves all messages in a conversation
	GetConversationHistory(ctx context.Context, conversationID string, limit int) ([]*models.OttoMessage, error)

	// GenerateResponse generates a response using the AI model
	// Takes conversation history and system context for better responses
	GenerateResponse(ctx context.Context, messages []*models.OttoMessage, context *models.ConversationContext) (string, error)

	// EnrichContext adds relevant data to conversation context
	// E.g., recent media, analytics, documents
	EnrichContext(ctx context.Context, conversationID string) (*models.ConversationContext, error)
}

// OttoConversationManager handles conversation creation and management
type OttoConversationManager interface {
	// CreateConversation starts a new conversation with optional initial context
	CreateConversation(ctx context.Context, userID, workspaceID, title string, initialContext *models.ConversationContext) (*models.OttoConversation, error)

	// ListConversations retrieves paginated conversations
	ListConversations(ctx context.Context, userID, workspaceID string, limit, offset int) ([]*models.OttoConversation, error)

	// GetConversation retrieves a single conversation
	GetConversation(ctx context.Context, conversationID string) (*models.OttoConversation, error)

	// UpdateConversation updates conversation metadata
	UpdateConversation(ctx context.Context, conversationID string, updates *models.UpdateOttoConversationInput) error

	// ArchiveConversation moves conversation to archived status
	ArchiveConversation(ctx context.Context, conversationID string) error

	// BookmarkConversation toggles the bookmark status
	BookmarkConversation(ctx context.Context, conversationID string, isBookmarked bool) error

	// DeleteConversation removes a conversation (soft delete)
	DeleteConversation(ctx context.Context, conversationID string) error
}

// ConversationService is the concrete implementation of conversation operations
type ConversationService struct {
	conversationRepo repository.OttoConversationRepository
	messageRepo      repository.OttoMessageRepository
}

// NewConversationService creates a new conversation service
func NewConversationService(
	conversationRepo repository.OttoConversationRepository,
	messageRepo repository.OttoMessageRepository,
) OttoConversationManager {
	return &ConversationService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
	}
}

// CreateConversation implements OttoConversationManager
func (s *ConversationService) CreateConversation(
	ctx context.Context,
	userID, workspaceID, title string,
	initialContext *models.ConversationContext,
) (*models.OttoConversation, error) {
	// Serialize context to JSON
	contextJSON, _ := json.Marshal(initialContext)

	conversation := models.NewOttoConversation(
		userID,
		workspaceID,
		title,
		"",
		string(contextJSON),
	)

	if err := s.conversationRepo.CreateConversation(ctx, conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

// ListConversations implements OttoConversationManager
func (s *ConversationService) ListConversations(
	ctx context.Context,
	userID, workspaceID string,
	limit, offset int,
) ([]*models.OttoConversation, error) {
	return s.conversationRepo.ListConversations(ctx, userID, workspaceID, limit, offset)
}

// GetConversation implements OttoConversationManager
func (s *ConversationService) GetConversation(ctx context.Context, conversationID string) (*models.OttoConversation, error) {
	return s.conversationRepo.GetConversation(ctx, conversationID)
}

// UpdateConversation implements OttoConversationManager
func (s *ConversationService) UpdateConversation(
	ctx context.Context,
	conversationID string,
	updates *models.UpdateOttoConversationInput,
) error {
	conversation, err := s.conversationRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return err
	}

	if updates.Title != "" {
		conversation.Title = updates.Title
	}
	if updates.Description != "" {
		conversation.Description.String = updates.Description
		conversation.Description.Valid = true
	}

	conversation.IsBookmarked = updates.IsBookmarked

	return s.conversationRepo.UpdateConversation(ctx, conversation)
}

// ArchiveConversation implements OttoConversationManager
func (s *ConversationService) ArchiveConversation(ctx context.Context, conversationID string) error {
	conversation, err := s.conversationRepo.GetConversation(ctx, conversationID)
	if err != nil {
		return err
	}

	conversation.Status = "archived"
	return s.conversationRepo.UpdateConversation(ctx, conversation)
}

// BookmarkConversation implements OttoConversationManager
func (s *ConversationService) BookmarkConversation(ctx context.Context, conversationID string, isBookmarked bool) error {
	return s.conversationRepo.SetBookmarked(ctx, conversationID, isBookmarked)
}

// DeleteConversation implements OttoConversationManager
func (s *ConversationService) DeleteConversation(ctx context.Context, conversationID string) error {
	return s.conversationRepo.DeleteConversation(ctx, conversationID)
}
