package games

import (
	"context"

	"github.com/Germanicus1/kanban-sim/internal/models"
)

// GameService defines business-logic operations for games
type GameService interface {
	CreateGame(ctx context.Context, input CreateGameInput) (*models.Game, error)
	GetGame(ctx context.Context, id int64) (*models.Game, error)
	UpdateGame(ctx context.Context, id int64, input UpdateGameInput) (*models.Game, error)
	DeleteGame(ctx context.Context, id int64) error
	GetEvents(ctx context.Context, id int64) ([]Event, error)
	GetBoard(ctx context.Context, id int64) (*Board, error)
}

// gameService is the default implementation of GameService
type gameService struct {
	repo GameRepository
}

// NewGameService constructs a new GameService
func NewGameService(repo GameRepository) GameService {
	return &gameService{repo: repo}
}

func (s *gameService) CreateGame(ctx context.Context, input CreateGameInput) (*models.Game, error) {
	g := &models.Game{
		Name:      input.Name,
		CreatedAt: input.CreatedAt,
		// map additional fields
	}
	if err := s.repo.Create(ctx, g); err != nil {
		return nil, err
	}
	// any additional business logic (e.g., seeding effort types)
	return g, nil
}

// ...implement the rest of the service methods similarly
