package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type conversationRepositoryImpl struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewOttoConversationRepository creates a new Otto conversation repository
func NewOttoConversationRepository(db *sqlx.DB) OttoConversationRepository {
	return &conversationRepositoryImpl{
		db:     db,
		logger: logging.GetComponentLogger("otto_conversation_repository"),
	}
}

// CreateConversation creates a new conversation in the database
func (r *conversationRepositoryImpl) CreateConversation(ctx context.Context, conversation *models.OttoConversation) error {
	query := `
		INSERT INTO otto_conversations (
			id, user_id, workspace_id, title, description, context,
			status, is_bookmarked, message_count, last_message_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		conversation.ID,
		conversation.UserID,
		conversation.WorkspaceID,
		conversation.Title,
		conversation.Description,
		conversation.Context,
		conversation.Status,
		conversation.IsBookmarked,
		conversation.MessageCount,
		conversation.LastMessageAt,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create conversation",
			zap.String("conversation_id", conversation.ID),
			zap.String("user_id", conversation.UserID),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("Conversation created",
		zap.String("conversation_id", conversation.ID),
		zap.String("user_id", conversation.UserID),
	)

	return nil
}

// GetConversation retrieves a conversation by ID
func (r *conversationRepositoryImpl) GetConversation(ctx context.Context, conversationID string) (*models.OttoConversation, error) {
	query := `
		SELECT
			id, user_id, workspace_id, title, description, context,
			status, is_bookmarked, message_count, last_message_at, created_at, updated_at
		FROM otto_conversations
		WHERE id = $1 AND status != 'deleted'
	`

	var conversation models.OttoConversation
	err := r.db.GetContext(ctx, &conversation, query, conversationID)

	if err == sql.ErrNoRows {
		r.logger.Debug("Conversation not found",
			zap.String("conversation_id", conversationID),
		)
		return nil, sql.ErrNoRows
	}

	if err != nil {
		r.logger.Error("Failed to get conversation",
			zap.String("conversation_id", conversationID),
			zap.Error(err),
		)
		return nil, err
	}

	return &conversation, nil
}

// ListConversations retrieves conversations for a user/workspace with pagination
func (r *conversationRepositoryImpl) ListConversations(
	ctx context.Context,
	userID, workspaceID string,
	limit int, offset int,
) ([]*models.OttoConversation, error) {
	query := `
		SELECT
			id, user_id, workspace_id, title, description, context,
			status, is_bookmarked, message_count, last_message_at, created_at, updated_at
		FROM otto_conversations
		WHERE user_id = $1 AND workspace_id = $2 AND status != 'deleted'
		ORDER BY updated_at DESC
		LIMIT $3 OFFSET $4
	`

	var conversations []*models.OttoConversation
	err := r.db.SelectContext(ctx, &conversations, query, userID, workspaceID, limit, offset)

	if err != nil && err != sql.ErrNoRows {
		r.logger.Error("Failed to list conversations",
			zap.String("user_id", userID),
			zap.String("workspace_id", workspaceID),
			zap.Error(err),
		)
		return nil, err
	}

	if conversations == nil {
		conversations = make([]*models.OttoConversation, 0)
	}

	return conversations, nil
}

// UpdateConversation updates an existing conversation
func (r *conversationRepositoryImpl) UpdateConversation(ctx context.Context, conversation *models.OttoConversation) error {
	query := `
		UPDATE otto_conversations
		SET
			title = $1,
			description = $2,
			context = $3,
			status = $4,
			is_bookmarked = $5,
			message_count = $6,
			last_message_at = $7,
			updated_at = $8
		WHERE id = $9 AND status != 'deleted'
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		conversation.Title,
		conversation.Description,
		conversation.Context,
		conversation.Status,
		conversation.IsBookmarked,
		conversation.MessageCount,
		conversation.LastMessageAt,
		conversation.UpdatedAt,
		conversation.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update conversation",
			zap.String("conversation_id", conversation.ID),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No conversation found to update",
			zap.String("conversation_id", conversation.ID),
		)
		return sql.ErrNoRows
	}

	r.logger.Info("Conversation updated",
		zap.String("conversation_id", conversation.ID),
	)

	return nil
}

// ArchiveConversation marks a conversation as archived
func (r *conversationRepositoryImpl) ArchiveConversation(ctx context.Context, conversationID string) error {
	query := `
		UPDATE otto_conversations
		SET status = 'archived', updated_at = NOW()
		WHERE id = $1 AND status != 'deleted'
	`

	result, err := r.db.ExecContext(ctx, query, conversationID)
	if err != nil {
		r.logger.Error("Failed to archive conversation",
			zap.String("conversation_id", conversationID),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No conversation found to archive",
			zap.String("conversation_id", conversationID),
		)
		return sql.ErrNoRows
	}

	r.logger.Info("Conversation archived",
		zap.String("conversation_id", conversationID),
	)

	return nil
}

// DeleteConversation soft deletes a conversation
func (r *conversationRepositoryImpl) DeleteConversation(ctx context.Context, conversationID string) error {
	query := `
		UPDATE otto_conversations
		SET status = 'deleted', updated_at = NOW()
		WHERE id = $1 AND status != 'deleted'
	`

	result, err := r.db.ExecContext(ctx, query, conversationID)
	if err != nil {
		r.logger.Error("Failed to delete conversation",
			zap.String("conversation_id", conversationID),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No conversation found to delete",
			zap.String("conversation_id", conversationID),
		)
		return sql.ErrNoRows
	}

	r.logger.Info("Conversation deleted",
		zap.String("conversation_id", conversationID),
	)

	return nil
}

// SetBookmarked toggles the bookmark status of a conversation
func (r *conversationRepositoryImpl) SetBookmarked(ctx context.Context, conversationID string, isBookmarked bool) error {
	query := `
		UPDATE otto_conversations
		SET is_bookmarked = $1, updated_at = NOW()
		WHERE id = $2 AND status != 'deleted'
	`

	result, err := r.db.ExecContext(ctx, query, isBookmarked, conversationID)
	if err != nil {
		r.logger.Error("Failed to update bookmark status",
			zap.String("conversation_id", conversationID),
			zap.Bool("is_bookmarked", isBookmarked),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No conversation found to update bookmark",
			zap.String("conversation_id", conversationID),
		)
		return sql.ErrNoRows
	}

	r.logger.Info("Bookmark status updated",
		zap.String("conversation_id", conversationID),
		zap.Bool("is_bookmarked", isBookmarked),
	)

	return nil
}
