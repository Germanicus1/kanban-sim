package players

import (
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type PlayerRepository interface {
	CreatePlayer(name string) (*models.Player, error)
	GetPlayer(id uuid.UUID) (*models.Player, error)
	UpdatePlayer(id uuid.UUID, name string) (*models.Player, error)
	DeletePlayer(id uuid.UUID) error
	ListPlayersByGameID() ([]*models.Player, error)
}
