package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Column struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	WIPLimit   int      `json:"wipLimit,omitempty"`
	Subcolumns []Column `json:"subcolumns,omitempty"`
}

type Card struct {
	ID             string `json:"id"`
	ClassOfService string `json:"classOfService"`
	ColumnID       string `json:"columnId"`
	ValueEstimate  string `json:"valueEstimate"`
	Effort         struct {
		Analysis    int `json:"analysis"`
		Development int `json:"development"`
		Test        int `json:"test"`
	} `json:"effort"`
	SelectedDay int  `json:"selectedDay,omitempty"`
	DeployedDay *int `json:"deployedDay,omitempty"`
}

type BoardConfig struct {
	Columns      []Column `json:"columns"`
	InitialCards []Card   `json:"initialCards"`
}

// LoadConfig reads and parses the configuration JSON file
func LoadConfig(filePath string) (*BoardConfig, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config BoardConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config JSON: %w", err)
	}

	return &config, nil
}
