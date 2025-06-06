package config

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
)

//go:embed board_config.json
var boardFS embed.FS

// LoadBoardConfig loads the board configuration from the embedded file system.
// It returns a pointer to a models.Board struct and an error if any occurs.
func LoadBoardConfig() (*models.Board, error) {
	b, err := boardFS.ReadFile("board_config.json")
	if err != nil {
		return nil, fmt.Errorf("embed read failed: %w", err)
	}
	var cfg models.Board
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return &cfg, nil
}
