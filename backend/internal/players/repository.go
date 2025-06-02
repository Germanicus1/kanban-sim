package players

import (
	"context"
	"database/sql"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type Repository interface {
	CreatePlayer(ctx context.Context, gamid uuid.UUID, name string) (uuid.UUID, error)
	GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.Player, error)
	UpdatePlayer(ctx context.Context, id uuid.UUID, name string) error
	DeletePlayer(ctx context.Context, id uuid.UUID) error
	ListPlayersByGameID(ctx context.Context, gameID uuid.UUID) ([]*models.Player, error)
}

// sqlRepo implements the Repository interface using a SQL database.
func NewSQLRepo(db *sql.DB) Repository {
	return &sqlRepo{db: db}
}
