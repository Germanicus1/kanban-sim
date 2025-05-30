package games

import (
	"context"
	"database/sql"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

// Repository declares what data operations you need.
type Repository interface {
	CreateGame(ctx context.Context, cfg models.BoardConfig) (uuid.UUID, error)
	GetBoard(ctx context.Context, id uuid.UUID) (models.Board, error)
	GetGameByID(ctx context.Context, id uuid.UUID) (models.Game, error)
	DeleteGame(ctx context.Context, id uuid.UUID) error
	UpdateGame(ctx context.Context, id uuid.UUID, day int) error
	ListGames(ctx context.Context) ([]models.Game, error)
}

// NewSQLRepo constructs a games.Repository backed by *sql.DB.
func NewSQLRepo(db *sql.DB) Repository {
	return &sqlRepo{db: db}
}
