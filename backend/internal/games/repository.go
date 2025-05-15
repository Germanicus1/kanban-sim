package games

import (
	"context"
	"database/sql"

	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

// Repository declares what data operations you need.
type Repository interface {
	CreateGame(ctx context.Context, cfg models.BoardConfig) (uuid.UUID, error)
	GetBoard(ctx context.Context, gameID uuid.UUID) (models.Board, error)
	GetGameByID(ctx context.Context, id uuid.UUID) (models.Game, error)
}

// NewSQLRepo constructs a games.Repository backed by *sql.DB.
func NewSQLRepo(db *sql.DB) Repository {
	return &sqlRepo{db: db}
}
