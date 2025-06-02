package models

import "github.com/google/uuid"

type Player struct {
	ID     uuid.UUID
	Name   string
	GameID uuid.UUID
}

// CreatePlayerRequest is the payload for CreatePlayer.
// swagger:model
type CreatePlayerRequest struct {
	GameID uuid.UUID `json:"game_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name   string    `json:"name" example:"John"`
}

// swagger:model
type UpdatePlayerRequest struct {
	ID   uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name string    `json:"name" example:"John"`
}

type DeletePlayerRequest struct {
	ID uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
}
