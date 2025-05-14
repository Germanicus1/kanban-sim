package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/response"
	"github.com/google/uuid"
)

// GetEvents returns the audit log for a game.
func GetEvents(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.URL.Query().Get("game_id"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	rows, err := database.DB.Query(
		`SELECT id, card_id, event_type, payload, created_at 
           FROM game_events 
          WHERE game_id=$1 
       ORDER BY created_at`,
		gameID,
	)
	if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}
	defer rows.Close()

	var events []any
	for rows.Next() {
		var (
			id        uuid.UUID
			cardID    uuid.UUID
			typ       string
			raw       []byte
			createdAt time.Time
		)
		rows.Scan(&id, &cardID, &typ, &raw, &createdAt)
		events = append(events, map[string]any{
			"id":         id,
			"card_id":    cardID,
			"event_type": typ,
			"payload":    json.RawMessage(raw),
			"created_at": createdAt,
		})
	}
	response.RespondWithData(w, events)
}
