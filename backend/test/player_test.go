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

func TestPlayerCRUD(t *testing.T) {
	setupDB(t, "players")
	setupDB(t, "games")
	defer tearDownDB()

	// First, create a game to associate with the player
	var gameID uuid.UUID
	var playerID uuid.UUID
	t.Run("Create Game for Player", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/games", nil)
		w := httptest.NewRecorder()

		handlers.CreateGame(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		var response internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf("Expected success, got error: %s", response.Error)
		}

		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data object in response")
		}

		gameIDStr, _ := data["id"].(string)
		if gameIDStr == "" {
			t.Fatal("Expected a valid game ID")
		}

		gameID, _ = uuid.Parse(gameIDStr)
		t.Logf("Created Game ID for Player: %s", gameIDStr)
	})

	// --- Create Player ---
	t.Run("Create Player", func(t *testing.T) {
		playerData := map[string]interface{}{
			"game_id": gameID.String(),
			"name":    "Test Player",
		}
		body, _ := json.Marshal(playerData)

		t.Logf("Payload for Create Player: %s", string(body))

		req := httptest.NewRequest(http.MethodPost, "/players", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreatePlayer(w, req)

		res := w.Result()
		defer res.Body.Close()

		var response internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d, error: %s", res.StatusCode, response.Error)
		}

		if !response.Success {
			t.Fatalf("Expected success, got error: %s", response.Error)
		}

		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data object in response")
		}

		playerIDStr, _ := data["id"].(string)
		if playerIDStr == "" {
			t.Fatal("Expected a valid player ID")
		}

		playerID, _ = uuid.Parse(playerIDStr)
		t.Logf("Created Player ID: %s", playerIDStr)
	})

	// --- Get Player ---
	t.Run("Get Player", func(t *testing.T) {
		if playerID == uuid.Nil {
			t.Fatal("Player ID is nil. Create_Player test may have failed.")
		}

		t.Logf("Fetching Player with ID: %s", playerID)

		req := httptest.NewRequest(http.MethodGet, "/players/get?id="+playerID.String(), nil)
		w := httptest.NewRecorder()

		handlers.GetPlayer(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		var response internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if !response.Success {
			t.Fatalf("Expected success, got error: %s", response.Error)
		}
	})

	// --- Update Player ---
	t.Run("Update Player", func(t *testing.T) {
		if playerID == uuid.Nil {
			t.Fatal("Player ID is nil. Create_Player test may have failed.")
		}

		updateData := map[string]string{
			"name": "Jane Doe",
		}
		body, _ := json.Marshal(updateData)

		t.Logf("Payload for Update Player: %s", string(body))

		req := httptest.NewRequest(http.MethodPut, "/players/update?id="+playerID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.UpdatePlayer(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected status 204, got %d", res.StatusCode)
		}

		// Verify the update by fetching the player
		req = httptest.NewRequest(http.MethodGet, "/players/get?id="+playerID.String(), nil)
		w = httptest.NewRecorder()

		handlers.GetPlayer(w, req)

		res = w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		var response internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data object in response")
		}

		name, _ := data["name"].(string)
		if name != "Jane Doe" {
			t.Fatalf("Expected name 'Jane Doe', got '%s'", name)
		}

		t.Logf("Player name successfully updated to: %s", name)
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
