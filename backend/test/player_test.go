package test

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

func setupPlayerTestDB(t *testing.T) {
	setupEnv(t)

	var err error
	db, err = internal.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatalf("Database connection failed: %v", err)
	}

	_, err = db.Exec("DELETE FROM players")
	if err != nil {
		t.Fatalf("Failed to clean players table: %v", err)
	}

	_, err = db.Exec("DELETE FROM games")
	if err != nil {
		t.Fatalf("Failed to clean games table: %v", err)
	}
}

func TestPlayerCRUD(t *testing.T) {
	setupPlayerTestDB(t)
	defer tearDownDB()

	// First, create a game to associate with the player
	var gameID uuid.UUID
	t.Run("Create Game for Player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/games", nil)
		w := httptest.NewRecorder()

		handlers.CreateGame(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", res.StatusCode)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		idStr, ok := response["id"].(string)
		if !ok || idStr == "" {
			t.Fatal("Expected a valid 'id' in response")
		}

		gameID, _ = uuid.Parse(idStr)
		if gameID == uuid.Nil {
			t.Fatal("Expected valid UUID, got Nil")
		}
	})

	var playerID uuid.UUID

	// --- Create Player ---
	t.Run("Create Player", func(t *testing.T) {
		playerData := map[string]interface{}{
			"game_id": gameID.String(),
			"name":    "John Doe",
		}
		body, _ := json.Marshal(playerData)

		req := httptest.NewRequest(http.MethodPost, "/players", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreatePlayer(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", res.StatusCode)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		idStr, ok := response["id"].(string)
		if !ok || idStr == "" {
			t.Fatal("Expected a valid 'id' in response")
		}

		playerID, _ = uuid.Parse(idStr)
		if playerID == uuid.Nil {
			t.Fatal("Expected valid UUID, got Nil")
		}
	})

	// --- Get Player ---
	t.Run("Get Player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/players/get?id="+playerID.String(), nil)
		w := httptest.NewRecorder()

		handlers.GetPlayer(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		var player handlers.Player
		if err := json.NewDecoder(res.Body).Decode(&player); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if player.ID != playerID {
			t.Fatalf("Expected player ID %s, got %s", playerID, player.ID)
		}
	})

	// --- Update Player ---
	t.Run("Update Player", func(t *testing.T) {
		updateData := map[string]string{
			"name": "Jane Doe",
		}
		body, _ := json.Marshal(updateData)

		req := httptest.NewRequest(http.MethodPut, "/players/update?id="+playerID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.UpdatePlayer(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected status 204, got %d", res.StatusCode)
		}

		// Verify the update
		req = httptest.NewRequest(http.MethodGet, "/players/get?id="+playerID.String(), nil)
		w = httptest.NewRecorder()
		handlers.GetPlayer(w, req)

		res = w.Result()
		defer res.Body.Close()

		var player handlers.Player
		if err := json.NewDecoder(res.Body).Decode(&player); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if player.Name != "Jane Doe" {
			t.Fatalf("Expected name 'Jane Doe', got '%s'", player.Name)
		}
	})

	// --- Delete Player ---
	t.Run("Delete Player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/players/delete?id="+playerID.String(), nil)
		w := httptest.NewRecorder()

		handlers.DeletePlayer(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected status 204, got %d", res.StatusCode)
		}

		// Verify deletion
		req = httptest.NewRequest(http.MethodGet, "/players/get?id="+playerID.String(), nil)
		w = httptest.NewRecorder()
		handlers.GetPlayer(w, req)

		res = w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", res.StatusCode)
		}
	})
}
