package games

import (
	"context"

	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

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
