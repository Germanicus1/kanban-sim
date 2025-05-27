package games

import (
	"context"

	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

// ServiceInterface declares all the business-logic operations
// your HTTP handlers will call.
type ServiceInterface interface {
	CreateGame(ctx context.Context, cfg models.BoardConfig) (uuid.UUID, error)
	GetBoard(ctx context.Context, gameID uuid.UUID) (models.Board, error)
	GetGame(ctx context.Context, id uuid.UUID) (models.Game, error)
	DeleteGame(ctx context.Context, id uuid.UUID) error
	UpdateGame(ctx context.Context, id uuid.UUID, day int) error
}

// Service holds your business-logic methods.
type Service struct {
	repo Repository
}

// NewService constructs a Service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateGame calls into your repo to persist a new game and seed all data.
func (s *Service) CreateGame(ctx context.Context, cfg models.BoardConfig) (uuid.UUID, error) {
	return s.repo.CreateGame(ctx, cfg)
}

// GetBoard retrieves the full board for a given game.
func (s *Service) GetBoard(ctx context.Context, gameID uuid.UUID) (models.Board, error) {
	return s.repo.GetBoard(ctx, gameID)
}

// GetGame retrieves a game by its ID.
func (s *Service) GetGame(ctx context.Context, id uuid.UUID) (models.Game, error) {
	return s.repo.GetGameByID(ctx, id)
}

// DeleteGame forwards to the repository.
func (s *Service) DeleteGame(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteGame(ctx, id)
}

// UpdateGame forwards to the repository.
func (s *Service) UpdateGame(ctx context.Context, id uuid.UUID, day int) error {
	return s.repo.UpdateGame(ctx, id, day)
}
