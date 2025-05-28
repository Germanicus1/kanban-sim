package players

import (
	"context"

	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

type ServiceInterface interface {
	CreatePlayer(name string) (*models.Player, error)
	GetPlayer(id uuid.UUID) (*models.Player, error)
	UpdatePlayer(id uuid.UUID, name string) (*models.Player, error)
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

//TODO implement the methods below

func (s *Service) GetPlayer(id string) (*models.Player, error) {
	// TODO: implement
	panic("not implemented")
}

func (s *Service) UpdatePlayer(id, name string) (*models.Player, error) {
	// TODO: implement
	panic("not implemented")
}

func (s *Service) DeletePlayer(id uuid.UUID) error {
	// TODO: implement
	panic("not implemented")
}

func (s *Service) ListPlayers() ([]*models.Player, error) {
	// TODO: implement
	panic("not implemented")
}
