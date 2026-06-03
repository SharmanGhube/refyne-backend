package repository

import (
	"context"

	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
)

type OttoRepository interface {
	GetSettings(ctx context.Context, userID string) (*models.OttoSettings, error)
	UpsertSettings(ctx context.Context, userID string, input *models.UpdateOttoSettingsInput) (*models.OttoSettings, error)
}
