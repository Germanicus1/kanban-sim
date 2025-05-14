package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type GameEvent struct {
	ID        uuid.UUID       `db:"id"`
	GameID    uuid.UUID       `db:"game_id"`
	CardID    uuid.UUID       `db:"card_id"`
	EventType string          `db:"event_type"`
	Payload   json.RawMessage `db:"payload"`
	CreatedAt time.Time       `db:"created_at"`
}
