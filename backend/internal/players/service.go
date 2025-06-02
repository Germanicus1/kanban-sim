package players

import (
	"context"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type ServiceInterface interface {
	CreatePlayer(ctx context.Context, gamid uuid.UUID, name string) (uuid.UUID, error)
	GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.Player, error)
	UpdatePlayer(ctx context.Context, id uuid.UUID, name string) error
	DeletePlayer(ctx context.Context, id uuid.UUID) error
	ListPlayersByGameID(ctx context.Context, gameID uuid.UUID) ([]*models.Player, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePlayer(ctx context.Context, gameID uuid.UUID, name string) (uuid.UUID, error) {
	return s.repo.CreatePlayer(ctx, gameID, name)
}

func (s *Service) GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
	return s.repo.GetPlayerByID(ctx, id)
}

func (s *Service) UpdatePlayer(ctx context.Context, id uuid.UUID, name string) error {
	return s.repo.UpdatePlayer(ctx, id, name)
}

func (s *Service) DeletePlayer(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePlayer(ctx, id)
}

func (s *Service) ListPlayersByGameID(ctx context.Context, gameID uuid.UUID) ([]*models.Player, error) {
	return s.repo.ListPlayersByGameID(context.Background(), uuid.Nil) // Assuming gameID is not needed here
}
