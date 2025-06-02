package columns

import (
	"context"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type ColumnRepository interface {
	GetColumnsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Column, error)
}
