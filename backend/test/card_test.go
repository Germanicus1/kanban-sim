package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/handlers"
	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
)

func TestCardCRUD(t *testing.T) {
	setupDB(t, "cards")
	setupDB(t, "games")
	defer tearDownDB()

	var gameID, cardID uuid.UUID

	// Create a game first
	t.Run("Create Game for Card", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/games", nil)
		w := httptest.NewRecorder()
		handlers.CreateGame(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		json.NewDecoder(res.Body).Decode(&resp)
		m := resp.Data.(map[string]interface{})
		idStr := m["id"].(string)
		gameID, _ = uuid.Parse(idStr)
	})

	// CreateCard
	t.Run("Create Card", func(t *testing.T) {
		payload := map[string]interface{}{
			"game_id":     gameID.String(),
			"title":       "Test Card",
			"card_column": "options",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/cards", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handlers.CreateCard(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		json.NewDecoder(res.Body).Decode(&resp)
		if !resp.Success {
			t.Fatal("Expected success")
		}
		m := resp.Data.(map[string]interface{})
		cardID, _ = uuid.Parse(m["id"].(string))
	})

	// GetCard
	t.Run("Get Card", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		w := httptest.NewRecorder()
		handlers.GetCard(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		json.NewDecoder(res.Body).Decode(&resp)
		if !resp.Success {
			t.Fatal("Expected success")
		}
	})

	// UpdateCard
	t.Run("Update Card", func(t *testing.T) {
		payload := map[string]interface{}{
			"title":       "Updated",
			"card_column": "selected",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPut, "/cards/update?id="+cardID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handlers.UpdateCard(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected 204, got %d", res.StatusCode)
		}
	})

	// DeleteCard
	t.Run("Delete Card", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/cards/delete?id="+cardID.String(), nil)
		w := httptest.NewRecorder()
		handlers.DeleteCard(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected 204, got %d", res.StatusCode)
		}
	})

	// GetDeletedCard
	t.Run("Get Deleted Card", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		w := httptest.NewRecorder()
		handlers.GetCard(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		json.NewDecoder(res.Body).Decode(&resp)
		if resp.Error != internal.ErrCardNotFound {
			t.Fatalf("Expected error %q, got %q", internal.ErrCardNotFound, resp.Error)
		}
	})
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

	// 1) Verify current state
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

	// 2) Update the card’s column (and selected_day if moving into “selected”)
	_, err = internal.DB.Exec(
		`UPDATE cards SET card_column=$1, selected_day=$2 WHERE id=$3`,
		p.ToColumn, p.Day, cardID,
	)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}

	// 3) Log the event
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
