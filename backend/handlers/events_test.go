package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Germanicus1/kanban-sim/handlers"
	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
)

func TestGetEvents(t *testing.T) {
	// Clean up relevant tables
	setupDB(t, "game_events")
	setupDB(t, "cards")
	setupDB(t, "games")
	defer tearDownDB()

	var gameID, cardID uuid.UUID

	// 1) Create a game
	t.Run("Create Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/games", nil)
		w := httptest.NewRecorder()
		handlers.CreateGame(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("CreateGame: expected 200, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		m := resp.Data.(map[string]interface{})
		var err error
		gameID, err = uuid.Parse(m["id"].(string))
		if err != nil {
			t.Fatal(err)
		}
	})

	// 2) Create a card in "todo"
	t.Run("Create Card", func(t *testing.T) {
		payload := map[string]interface{}{
			"game_id":     gameID.String(),
			"title":       "Card for Events",
			"card_column": "todo",
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/cards", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handlers.CreateCard(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("CreateCard: expected 200, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		m := resp.Data.(map[string]interface{})
		var err error
		cardID, err = uuid.Parse(m["id"].(string))
		if err != nil {
			t.Fatal(err)
		}
	})

	// 3) Perform two moves to generate events
	moves := []map[string]interface{}{
		{"from_column": "todo", "to_column": "doing", "day": 1},
		{"from_column": "doing", "to_column": "done", "day": 2},
	}
	for i, mv := range moves {
		t.Run(fmt.Sprintf("MoveCard#%c", 'A'+i), func(t *testing.T) {
			body, _ := json.Marshal(mv)
			req := httptest.NewRequest(http.MethodPut, "/cards/move?id="+cardID.String(), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handlers.MoveCard(w, req)
			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != http.StatusNoContent {
				t.Fatalf("MoveCard #%d: expected 204, got %d", i, res.StatusCode)
			}
			// slight pause so created_at ordering is consistent
			time.Sleep(10 * time.Millisecond)
		})
	}

	// 4) Fetch events for the game
	req := httptest.NewRequest(http.MethodGet, "/games/events?game_id="+gameID.String(), nil)
	w := httptest.NewRecorder()
	handlers.GetEvents(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("GetEvents: expected 200, got %d", res.StatusCode)
	}

	var resp internal.APIResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatal("decoding response:", err)
	}
	if !resp.Success {
		t.Fatalf("GetEvents: expected success, got error %s", resp.Error)
	}

	events, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("GetEvents: expected data to be []interface{}")
	}
	if len(events) != len(moves) {
		t.Fatalf("GetEvents: expected %d events, got %d", len(moves), len(events))
	}

	// Validate each event
	for i, e := range events {
		ev := e.(map[string]interface{})
		if ev["event_type"] != "move" {
			t.Errorf("event[%d]: expected type 'move', got %v", i, ev["event_type"])
		}
		// payload is nested JSON
		pl := ev["payload"].(map[string]interface{})
		if pl["from"].(string) != moves[i]["from_column"] {
			t.Errorf("event[%d] payload.from: expected %v, got %v", i, moves[i]["from_column"], pl["from"])
		}
		if pl["to"].(string) != moves[i]["to_column"] {
			t.Errorf("event[%d] payload.to: expected %v, got %v", i, moves[i]["to_column"], pl["to"])
		}
		if int(pl["day"].(float64)) != moves[i]["day"].(int) {
			t.Errorf("event[%d] payload.day: expected %v, got %v", i, moves[i]["day"], pl["day"])
		}
		if ev["card_id"].(string) != cardID.String() {
			t.Errorf("event[%d] card_id: expected %s, got %v", i, cardID, ev["card_id"])
		}
	}
}
