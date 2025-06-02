package columns

import (
	"context"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type ColumnServiceInterface interface {
	GetColumnsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Column, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetColumnsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Column, error) {
	return s.repo.GetColumnsByGameID(ctx, gameID)
}
