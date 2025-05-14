// internal/config/types.go
package config

import "github.com/Germanicus1/kanban-sim/internal/models"

type ConfigCard struct {
	Title          string          `json:"title"`
	ColumnTitle    string          `json:"columnTitle"`
	ClassOfService *string         `json:"classOfService"`
	ValueEstimate  *string         `json:"valueEstimate"`
	SelectedDay    *int            `json:"selectedDay"`
	DeployedDay    *int            `json:"deployedDay"`
	Efforts        []models.Effort `json:"efforts"`
}
