package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadBoardConfig(path string) (*BoardConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read board config: %w", err)
	}

	var config BoardConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse board config JSON: %w", err)
	}

	return &config, nil
}
