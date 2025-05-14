package config

import (
	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

// Board bundles the full state of one game’s board.
type Board struct {
	GameID      uuid.UUID           `json:"gameId"`
	Columns     []models.Column     `json:"columns"`
	EffortTypes []models.EffortType `json:"effortTypes"`
	Cards       []ConfigCard        `json:"cards"`
}
