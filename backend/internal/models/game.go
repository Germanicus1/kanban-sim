package models

import (
	"github.com/google/uuid"
)

type Game struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt string    `json:"created_at"`
	Day       int       `json:"day"`
}
