package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
)

type ottoRepositoryImpl struct {
	db *sqlx.DB
}

func NewOttoRepository(db *sqlx.DB) OttoRepository {
	return &ottoRepositoryImpl{db: db}
}

func (r *ottoRepositoryImpl) GetSettings(ctx context.Context, userID string) (*models.OttoSettings, error) {
	query := `SELECT * FROM otto_settings WHERE user_id = $1`
	var settings models.OttoSettings
	err := r.db.GetContext(ctx, &settings, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return default settings if none exist
			return &models.OttoSettings{
				UserID:             userID,
				AutoReplyEnabled:   false,
				Tone:               "professional",
				Guardrails:         "",
				CustomInstructions: "",
			}, nil
		}
		return nil, err
	}
	return &settings, nil
}

func (r *ottoRepositoryImpl) UpsertSettings(ctx context.Context, userID string, input *models.UpdateOttoSettingsInput) (*models.OttoSettings, error) {
	query := `
		INSERT INTO otto_settings (user_id, auto_reply_enabled, tone, guardrails, custom_instructions)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			auto_reply_enabled = EXCLUDED.auto_reply_enabled,
			tone = EXCLUDED.tone,
			guardrails = EXCLUDED.guardrails,
			custom_instructions = EXCLUDED.custom_instructions
		RETURNING *
	`
	var settings models.OttoSettings
	err := r.db.GetContext(ctx, &settings, query, userID, input.AutoReplyEnabled, input.Tone, input.Guardrails, input.CustomInstructions)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}
