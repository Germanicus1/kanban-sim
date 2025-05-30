package players

import (
	"context"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type ServiceInterface interface {
	CreatePlayer(name string) (*models.Player, error)
	GetPlayerByID(id uuid.UUID) (*models.Player, error)
	UpdatePlayer(id uuid.UUID, name string) error
	DeletePlayer(id uuid.UUID) error
	ListPlayers() ([]*models.Player, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePlayer(ctx context.Context, cfg models.Player) (uuid.UUID, error) {
	return s.repo.CreatePlayer(ctx, cfg)
}

func (s *Service) GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
	return s.repo.GetPlayerByID(ctx, id)
}

// TODO implement the methods below
func (s *Service) UpdatePlayer(ctx context.Context, id uuid.UUID, name string) error {
	return s.repo.UpdatePlayer(ctx, id, name)
}

func (s *Service) DeletePlayer(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePlayer(ctx, id)
}

func (s *Service) ListPlayers(ctx context.Context) ([]*models.Player, error) {
	return s.repo.ListPlayers(context.Background(), uuid.Nil) // Assuming gameID is not needed here
}
