package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/domains/instagram/models"
)

type InstagramCommentRepository interface {
	SaveComment(ctx context.Context, comment *models.InstagramComment) error
	GetCommentsByAccount(ctx context.Context, accountID string) ([]*models.InstagramComment, error)
	UpdateCommentStatus(ctx context.Context, commentID string, isHidden bool) error
}

type instagramCommentRepository struct {
	db *sqlx.DB
}

func NewInstagramCommentRepository(db *sqlx.DB) InstagramCommentRepository {
	return &instagramCommentRepository{
		db: db,
	}
}

func (r *instagramCommentRepository) SaveComment(ctx context.Context, comment *models.InstagramComment) error {
	query := `
		INSERT INTO instagram_comments (
			id, instagram_media_id, account_id, username, text, is_hidden, is_flagged, moderation_reason, created_at, updated_at
		) VALUES (
			:id, :instagram_media_id, :account_id, :username, :text, :is_hidden, :is_flagged, :moderation_reason, :created_at, :updated_at
		)
		ON CONFLICT (id) DO UPDATE SET
			is_hidden = EXCLUDED.is_hidden,
			is_flagged = EXCLUDED.is_flagged,
			moderation_reason = EXCLUDED.moderation_reason,
			updated_at = NOW()
	`
	
	_, err := r.db.NamedExecContext(ctx, query, comment)
	if err != nil {
		return fmt.Errorf("failed to save comment: %w", err)
	}

	return nil
}

func (r *instagramCommentRepository) GetCommentsByAccount(ctx context.Context, accountID string) ([]*models.InstagramComment, error) {
	query := `
		SELECT id, instagram_media_id, account_id, username, text, is_hidden, is_flagged, moderation_reason, created_at, updated_at
		FROM instagram_comments
		WHERE account_id = $1
		ORDER BY created_at DESC
	`
	
	var comments []*models.InstagramComment
	err := r.db.SelectContext(ctx, &comments, query, accountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.InstagramComment{}, nil
		}
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	
	return comments, nil
}

func (r *instagramCommentRepository) UpdateCommentStatus(ctx context.Context, commentID string, isHidden bool) error {
	query := `
		UPDATE instagram_comments
		SET is_hidden = $1, updated_at = NOW()
		WHERE id = $2
	`
	
	_, err := r.db.ExecContext(ctx, query, isHidden, commentID)
	if err != nil {
		return fmt.Errorf("failed to update comment status: %w", err)
	}
	
	return nil
}
