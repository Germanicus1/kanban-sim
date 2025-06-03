package cards

import (
	"context"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type CardsServiceInterface interface {
	GetCardsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Card, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCardsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Card, error) {
	return s.repo.GetCardsByGameID(ctx, gameID)
}
