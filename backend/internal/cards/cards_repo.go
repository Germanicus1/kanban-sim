package cards

import (
	"context"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type Cardrepository interface {
	GetCardsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Card, error)
}
