package games

import (
	"context"
	"errors"

	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

// GameRepository defines data-access methods for games
type GameRepository interface {
	Create(ctx context.Context, g *models.Game) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Game, error)
	Update(ctx context.Context, g *models.Game) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListEvents(ctx context.Context, gameID uuid.UUID) ([]models.GameEvent, error)
	GetBoard(ctx context.Context, gameID uuid.UUID) (*models.Board, error)
}

type BoardRepository interface {
	// Fetches columns, cards, efforts, etc.
	GetBoard(ctx context.Context, gameID string) (map[string][]models.Card, error)
}

type CardRepository interface {
	Create(ctx context.Context, c *models.Card) error
	// …Get, Update, Delete…
}
type EventRepository interface {
	List(ctx context.Context, gameID string) ([]models.GameEvent, error)
}
