package players_test

import (
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/Germanicus1/kanban-sim/backend/internal/players"
	"github.com/google/uuid"
)

type Service interface {
	CreatePlayer(name string) (*models.Player, error)
	GetPlayer(id uuid.UUID) (*models.Player, error)
	UpdatePlayer(id uuid.UUID, name string) (*models.Player, error)
	DeletePlayer(id uuid.UUID) error
	ListPlayers() ([]*models.Player, error)
}

type service struct {
	repo players.Repository
}

func NewService(r players.Repository) Service {
	return &service{repo: r}
}

func (s *service) CreatePlayer(name string) (*models.Player, error) {
	// TODO: implement
	panic("not implemented")
}

func (s *service) GetPlayer(id uuid.UUID) (*models.Player, error) {
	// TODO: implement
	panic("not implemented")
}

func (s *service) UpdatePlayer(id uuid.UUID, name string) (*models.Player, error) {
	// TODO: implement
	panic("not implemented")
}

func (s *service) DeletePlayer(id uuid.UUID) error {
	// TODO: implement
	panic("not implemented")
}

func (s *service) ListPlayers() ([]*models.Player, error) {
	// TODO: implement
	panic("not implemented")
}
