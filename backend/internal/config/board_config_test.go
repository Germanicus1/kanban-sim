package config_test

import (
	"testing"

	"github.com/Germanicus1/kanban-sim/internal/config"
)

// TestLoadBoardConfig verifies that the embedded board_config.json
// is parsed correctly into a models.Board struct.
func TestLoadBoardConfig(t *testing.T) {
	cfg, err := config.LoadBoardConfig()
	if err != nil {
		t.Fatalf("LoadBoardConfig returned error: %v", err)
	}

	// There should be at least one card defined in the config
	if len(cfg.Cards) == 0 {
		t.Errorf("expected at least one card, got 0")
	}

	// Spot-check the first card matches the JSON example
	first := cfg.Cards[0]
	if first.Title != "S1" {
		t.Errorf("first card Title = %q; want %q", first.Title, "S1")
	}
	if len(first.Efforts) != 3 {
		t.Errorf("first card Efforts length = %d; want %d", len(first.Efforts), 3)
	}
	ef := first.Efforts[0]
	if ef.EffortType != "Analysis" || ef.Estimate != 4 {
		t.Errorf("first effort = %+v; want {EffortType:\"Analysis\", Estimate:4}", ef)
	}

	// The config should also define at least one effort type
	if len(cfg.EffortTypes) == 0 {
		t.Errorf("expected at least one EffortType, got 0")
	}
}
