package cards

import (
	"context"
	"database/sql"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type Repository interface {
	GetCardsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Card, error)
}

// sqlRepo implements the Repository interface using a SQL database.
func NewSQLRepo(db *sql.DB) Repository {
	return &sqlRepo{db: db}
}
