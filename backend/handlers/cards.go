package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
)

type Card struct {
	ID             uuid.UUID `json:"id"`
	GameID         uuid.UUID `json:"game_id"`
	Title          string    `json:"title"`
	CardColumn     string    `json:"card_column"`
	ClassOfService string    `json:"class_of_service,omitempty"`
	ValueEstimate  string    `json:"value_estimate,omitempty"`
	EffortAnalysis int       `json:"effort_analysis,omitempty"`
	EffortDev      int       `json:"effort_development,omitempty"`
	EffortTest     int       `json:"effort_test,omitempty"`
	SelectedDay    int       `json:"selected_day,omitempty"`
	DeployedDay    int       `json:"deployed_day,omitempty"`
}

// CreateCard creates a new card
func CreateCard(w http.ResponseWriter, r *http.Request) {
	var card Card
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrValidationFailed)
		return
	}
	if card.GameID == uuid.Nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidCardID)
		return
	}

	const q = `
        INSERT INTO cards (
            game_id, title, card_column,
            class_of_service, value_estimate,
            effort_analysis, effort_development, effort_test,
            selected_day, deployed_day
        ) VALUES (
            $1,$2,$3,$4,$5,$6,$7,$8,$9,$10
        ) RETURNING id
    `
	if err := internal.DB.QueryRow(
		q,
		card.GameID,
		card.Title,
		card.CardColumn,
		card.ClassOfService,
		card.ValueEstimate,
		card.EffortAnalysis,
		card.EffortDev,
		card.EffortTest,
		card.SelectedDay,
		card.DeployedDay,
	).Scan(&card.ID); err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}

	internal.RespondWithData(w, card)
}

// GetCard retrieves a card by ID
func GetCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidCardID)
		return
	}

	const q = `
        SELECT id, game_id, title, card_column
          FROM cards
         WHERE id = $1
    `
	var card Card
	err = internal.DB.QueryRow(q, id).Scan(
		&card.ID,
		&card.GameID,
		&card.Title,
		&card.CardColumn,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			internal.RespondWithError(w, http.StatusNotFound, internal.ErrCardNotFound)
		} else {
			status, code := internal.MapPostgresError(err)
			internal.RespondWithError(w, status, code)
		}
		return
	}

	internal.RespondWithData(w, card)
}

// UpdateCard updates card details
func UpdateCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidCardID)
		return
	}

	var card Card
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrValidationFailed)
		return
	}

	const q = `
        UPDATE cards SET
            title =$1,
            card_column =$2,
            class_of_service =$3,
            value_estimate =$4,
            effort_analysis =$5,
            effort_development =$6,
            effort_test =$7,
            selected_day =$8,
            deployed_day =$9
         WHERE id = $10
    `
	res, err := internal.DB.Exec(
		q,
		card.Title,
		card.CardColumn,
		card.ClassOfService,
		card.ValueEstimate,
		card.EffortAnalysis,
		card.EffortDev,
		card.EffortTest,
		card.SelectedDay,
		card.DeployedDay,
		id,
	)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		internal.RespondWithError(w, http.StatusNotFound, internal.ErrCardNotFound)
		return
	}

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// DeleteCard deletes a card by ID
func DeleteCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidCardID)
		return
	}

	const q = `DELETE FROM cards WHERE id = $1`
	res, err := internal.DB.Exec(q, id)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		internal.RespondWithError(w, http.StatusNotFound, internal.ErrCardNotFound)
		return
	}

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// MovePayload defines the body for moving a card.
type MovePayload struct {
	FromColumn string `json:"from_column"`
	ToColumn   string `json:"to_column"`
	Day        int    `json:"day"`
}

// MoveCard moves a card from one column to another and logs the event.
func MoveCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	cardID, err := uuid.Parse(idStr)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidCardID)
		return
	}

	var p MovePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrValidationFailed)
		return
	}

	// 1. Verify current state
	var current string
	err = internal.DB.QueryRow(
		`SELECT card_column FROM cards WHERE id = $1`, cardID,
	).Scan(&current)
	if err == sql.ErrNoRows {
		internal.RespondWithError(w, http.StatusNotFound, internal.ErrCardNotFound)
		return
	} else if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}
	if current != p.FromColumn {
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_MOVE_FROM")
		return
	}

	// 2. Update the card’s column (and selected_day if moving into “selected”)
	_, err = internal.DB.Exec(
		`UPDATE cards SET card_column=$1, selected_day=$2 WHERE id=$3`,
		p.ToColumn, p.Day, cardID,
	)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}

	// 3. Log the event
	payload := map[string]interface{}{
		"from": p.FromColumn,
		"to":   p.ToColumn,
		"day":  p.Day,
	}
	_, err = internal.DB.Exec(
		`INSERT INTO game_events (game_id, card_id, event_type, payload) 
           VALUES (
             (SELECT game_id FROM cards WHERE id=$1),
             $1, 'move', $2::jsonb
           )`,
		cardID,
		internal.ToJSON(payload),
	)
	if err != nil {
		// event‐logging failures shouldn’t block the move
		log.Printf("warning: failed to log move event: %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
