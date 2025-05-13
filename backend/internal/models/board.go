package models

import "github.com/google/uuid"

// Board bundles the full state of one game’s board.
type Board struct {
	GameID      uuid.UUID    `json:"gameId"`
	Columns     []Column     `json:"columns"`
	EffortTypes []EffortType `json:"effortTypes"`
	Cards       []Card       `json:"cards"`
}
