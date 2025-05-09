package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/handlers"
	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
)

func TestGetBoard(t *testing.T) {
	// Clean and prepare
	setupDB(t, "cards")
	setupDB(t, "games")
	defer tearDownDB()

	var gameID uuid.UUID

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
			t.Fatalf("CreateGame decode: %v", err)
		}
		m := resp.Data.(map[string]interface{})
		var err error
		gameID, err = uuid.Parse(m["id"].(string))
		if err != nil {
			t.Fatalf("CreateGame parse ID: %v", err)
		}
	})

	// 2) Seed cards in different columns
	cards := []struct {
		Title, Column string
	}{
		{"Card A", "todo"},
		{"Card B", "selected"},
		{"Card C", "done"},
	}
	for _, c := range cards {
		payload := map[string]interface{}{
			"game_id":     gameID.String(),
			"title":       c.Title,
			"card_column": c.Column,
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/cards", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handlers.CreateCard(w, req)
		res := w.Result()
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("CreateCard %q: expected 200, got %d", c.Title, res.StatusCode)
		}
	}

	// 3) Invoke GetBoard
	req := httptest.NewRequest(http.MethodGet, "/games/board?game_id="+gameID.String(), nil)
	w := httptest.NewRecorder()
	handlers.GetBoard(w, req)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("GetBoard: expected 200, got %d", res.StatusCode)
	}
	var resp internal.APIResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Fatalf("GetBoard decode: %v", err)
	}
	if !resp.Success {
		t.Fatalf("GetBoard: expected success, got error %q", resp.Error)
	}

	// 4) Assert cards appear under correct column keys
	board, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("GetBoard: expected map[string]interface{}, got %T", resp.Data)
	}

	for _, c := range cards {
		rawList, found := board[c.Column]
		if !found {
			t.Errorf("GetBoard: missing column %q", c.Column)
			continue
		}
		list := rawList.([]interface{})
		matched := false
		for _, item := range list {
			cardObj := item.(map[string]interface{})
			if cardObj["title"] == c.Title {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("GetBoard: card %q not found in column %q", c.Title, c.Column)
		}
	}
}
