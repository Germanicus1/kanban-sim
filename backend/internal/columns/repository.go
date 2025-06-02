package columns

import (
	"context"
	"database/sql"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type Repository interface {
	// GetColumnsByGameID retrieves all columns for a given game ID.
	GetColumnsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Column, error)
}

func NewSQLRepo(db *sql.DB) Repository {
	return &sqlRepo{db: db}
}
