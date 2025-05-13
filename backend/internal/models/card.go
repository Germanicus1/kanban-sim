package models

import "github.com/google/uuid"

// Card is your “clean” domain model for a Kanban card.
type Card struct {
	ID             uuid.UUID `json:"id"`
	GameID         uuid.UUID `json:"gameId"`
	ColumnID       uuid.UUID `json:"columnId"`
	Title          string    `json:"title"`
	ClassOfService *string   `json:"classOfService,omitempty"`
	ValueEstimate  *string   `json:"valueEstimate,omitempty"`
	SelectedDay    *int      `json:"selectedDay,omitempty"`
	DeployedDay    *int      `json:"deployedDay,omitempty"`
	Efforts        []Effort  `json:"efforts"`
}
