package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/Germanicus1/kanban-sim/internal/handlers"
	"github.com/google/uuid"
)

func TestPlayerCRUD(t *testing.T) {
	setupDB(t, "players")
	setupDB(t, "games")
	defer tearDownDB()

	var gameID, playerID uuid.UUID

	// Create a game for the player
	t.Run("Create Game for Player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/games", nil)
		w := httptest.NewRecorder()
		handlers.CreateGame(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		m := resp.Data.(map[string]interface{})
		gameID, _ = uuid.Parse(m["id"].(string))
	})

	// CreatePlayer
	t.Run("Create Player", func(t *testing.T) {
		payload := map[string]interface{}{
			"game_id": gameID.String(),
			"name":    "Test Player",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/players", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handlers.CreatePlayer(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if !resp.Success {
			t.Fatal("Expected success:", resp.Error)
		}
		m := resp.Data.(map[string]interface{})
		playerID, _ = uuid.Parse(m["id"].(string))
	})

	// GetPlayer
	t.Run("Get Player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/players/get?id="+playerID.String(), nil)
		w := httptest.NewRecorder()
		handlers.GetPlayer(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.StatusCode)
		}
		var resp internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		if !resp.Success {
			t.Fatal("Expected success:", resp.Error)
		}
	})

	// UpdatePlayer
	t.Run("Update Player", func(t *testing.T) {
		payload := map[string]interface{}{
			"name": "Jane Doe",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPut, "/players/update?id="+playerID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handlers.UpdatePlayer(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected 204, got %d", res.StatusCode)
		}

		// verify via GetPlayer
		req = httptest.NewRequest(http.MethodGet, "/players/get?id="+playerID.String(), nil)
		w = httptest.NewRecorder()
		handlers.GetPlayer(w, req)
		res = w.Result()
		defer res.Body.Close()

		var resp internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}
		m := resp.Data.(map[string]interface{})
		if m["name"] != "Jane Doe" {
			t.Fatalf("Expected name Jane Doe, got %v", m["name"])
		}
	})

	// DeletePlayer
	t.Run("Delete Player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/players/delete?id="+playerID.String(), nil)
		w := httptest.NewRecorder()
		handlers.DeletePlayer(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected 204, got %d", res.StatusCode)
		}

		// verify deletion
		req = httptest.NewRequest(http.MethodGet, "/players/get?id="+playerID.String(), nil)
		w = httptest.NewRecorder()
		handlers.GetPlayer(w, req)
		res = w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404, got %d", res.StatusCode)
		}
	})
}
