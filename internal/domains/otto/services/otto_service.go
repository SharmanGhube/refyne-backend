package services

import (
	"context"

	"github.com/refynehq/refyne-backend/internal/domains/otto/models"
	"github.com/refynehq/refyne-backend/internal/domains/otto/repository"
)

type OttoService interface {
	GetSettings(ctx context.Context, userID string) (*models.OttoSettings, error)
	UpsertSettings(ctx context.Context, userID string, input *models.UpdateOttoSettingsInput) (*models.OttoSettings, error)
}

type ottoServiceImpl struct {
	repo repository.OttoRepository
}

func NewOttoService(repo repository.OttoRepository) OttoService {
	return &ottoServiceImpl{repo: repo}
}

func (s *ottoServiceImpl) GetSettings(ctx context.Context, userID string) (*models.OttoSettings, error) {
	return s.repo.GetSettings(ctx, userID)
}

func (s *ottoServiceImpl) UpsertSettings(ctx context.Context, userID string, input *models.UpdateOttoSettingsInput) (*models.OttoSettings, error) {
	// Set default tone if empty
	if input.Tone == "" {
		input.Tone = "professional"
	}
	return s.repo.UpsertSettings(ctx, userID, input)
}
