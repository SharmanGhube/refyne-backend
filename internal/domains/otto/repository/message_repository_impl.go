package repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
	"github.com/refynehq/refyne-backend/pkg/logging"
	"go.uber.org/zap"
)

type messageRepositoryImpl struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewOttoMessageRepository creates a new Otto message repository
func NewOttoMessageRepository(db *sqlx.DB) OttoMessageRepository {
	return &messageRepositoryImpl{
		db:     db,
		logger: logging.GetComponentLogger("otto_message_repository"),
	}
}

// CreateMessage creates a new message in the database
func (r *messageRepositoryImpl) CreateMessage(ctx context.Context, message *models.OttoMessage) error {
	query := `
		INSERT INTO otto_messages (
			id, conversation_id, user_id, role, content,
			tokens_used, model_used, metadata, is_liked, feedback_notes, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		message.ID,
		message.ConversationID,
		message.UserID,
		message.Role,
		message.Content,
		message.TokensUsed,
		message.ModelUsed,
		message.Metadata,
		message.IsLiked,
		message.FeedbackNotes,
		message.CreatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create message",
			zap.String("message_id", message.ID),
			zap.String("conversation_id", message.ConversationID),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("Message created",
		zap.String("message_id", message.ID),
		zap.String("conversation_id", message.ConversationID),
	)

	return nil
}

// GetMessage retrieves a message by ID
func (r *messageRepositoryImpl) GetMessage(ctx context.Context, messageID string) (*models.OttoMessage, error) {
	query := `
		SELECT
			id, conversation_id, user_id, role, content,
			tokens_used, model_used, metadata, is_liked, feedback_notes, created_at
		FROM otto_messages
		WHERE id = $1
	`

	var message models.OttoMessage
	err := r.db.GetContext(ctx, &message, query, messageID)

	if err == sql.ErrNoRows {
		r.logger.Debug("Message not found",
			zap.String("message_id", messageID),
		)
		return nil, sql.ErrNoRows
	}

	if err != nil {
		r.logger.Error("Failed to get message",
			zap.String("message_id", messageID),
			zap.Error(err),
		)
		return nil, err
	}

	return &message, nil
}

// GetConversationMessages retrieves all messages for a conversation with pagination
func (r *messageRepositoryImpl) GetConversationMessages(
	ctx context.Context,
	conversationID string,
	limit int, offset int,
) ([]*models.OttoMessage, error) {
	query := `
		SELECT
			id, conversation_id, user_id, role, content,
			tokens_used, model_used, metadata, is_liked, feedback_notes, created_at
		FROM otto_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	var messages []*models.OttoMessage
	err := r.db.SelectContext(ctx, &messages, query, conversationID, limit, offset)

	if err != nil && err != sql.ErrNoRows {
		r.logger.Error("Failed to list messages",
			zap.String("conversation_id", conversationID),
			zap.Error(err),
		)
		return nil, err
	}

	if messages == nil {
		messages = make([]*models.OttoMessage, 0)
	}

	return messages, nil
}

// UpdateMessage updates an existing message
func (r *messageRepositoryImpl) UpdateMessage(ctx context.Context, message *models.OttoMessage) error {
	query := `
		UPDATE otto_messages
		SET
			role = $1,
			content = $2,
			tokens_used = $3,
			model_used = $4,
			metadata = $5,
			is_liked = $6,
			feedback_notes = $7
		WHERE id = $8
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		message.Role,
		message.Content,
		message.TokensUsed,
		message.ModelUsed,
		message.Metadata,
		message.IsLiked,
		message.FeedbackNotes,
		message.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update message",
			zap.String("message_id", message.ID),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No message found to update",
			zap.String("message_id", message.ID),
		)
		return sql.ErrNoRows
	}

	r.logger.Info("Message updated",
		zap.String("message_id", message.ID),
	)

	return nil
}

// AddFeedback adds or updates feedback for a message
func (r *messageRepositoryImpl) AddFeedback(ctx context.Context, messageID string, isLiked bool, notes string) error {
	query := `
		UPDATE otto_messages
		SET is_liked = $1, feedback_notes = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, isLiked, notes, messageID)
	if err != nil {
		r.logger.Error("Failed to add feedback",
			zap.String("message_id", messageID),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No message found to add feedback",
			zap.String("message_id", messageID),
		)
		return sql.ErrNoRows
	}

	r.logger.Info("Feedback added",
		zap.String("message_id", messageID),
		zap.Bool("is_liked", isLiked),
	)

	return nil
}

// GetMessageCount returns the total number of messages in a conversation
func (r *messageRepositoryImpl) GetMessageCount(ctx context.Context, conversationID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM otto_messages
		WHERE conversation_id = $1
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, conversationID)

	if err != nil {
		r.logger.Error("Failed to get message count",
			zap.String("conversation_id", conversationID),
			zap.Error(err),
		)
		return 0, err
	}

	return count, nil
}

// DeleteMessage soft deletes a message (marks as deleted, could also hard delete)
func (r *messageRepositoryImpl) DeleteMessage(ctx context.Context, messageID string) error {
	// For now, we do a hard delete since there's no deleted_at column on messages
	// In a full implementation, you might add a deleted_at column and soft delete instead
	query := `
		DELETE FROM otto_messages
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, messageID)
	if err != nil {
		r.logger.Error("Failed to delete message",
			zap.String("message_id", messageID),
			zap.Error(err),
		)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No message found to delete",
			zap.String("message_id", messageID),
		)
		return sql.ErrNoRows
	}

	r.logger.Info("Message deleted",
		zap.String("message_id", messageID),
	)

	return nil
}
