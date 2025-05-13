package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Game struct {
	ID        uuid.UUID       `json:"id"`
	CreatedAt string          `json:"created_at"`
	Day       int             `json:"day"`
	Columns   json.RawMessage `json:"columns"`
}
