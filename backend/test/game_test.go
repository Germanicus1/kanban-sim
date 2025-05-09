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

func TestGameCRUD(t *testing.T) {
	setupDB(t, "games")
	defer tearDownDB()
	var gameID uuid.UUID

	// --- Create Game ---
	t.Run("Create Game", func(t *testing.T) {
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

		idStr, ok := data["id"].(string)
		if !ok || idStr == "" {
			t.Fatal("Expected a valid 'id' in response")
		}

		gameID, _ = uuid.Parse(idStr)
		if gameID == uuid.Nil {
			t.Fatal("Expected valid UUID, got Nil")
		}
	})

	// --- Get Game ---
	t.Run("Get Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/games/get?id="+gameID.String(), nil)
		w := httptest.NewRecorder()

		handlers.GetGame(w, req)

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

		idStr, _ := data["id"].(string)
		if idStr != gameID.String() {
			t.Fatalf("Expected game ID %s, got %s", gameID.String(), idStr)
		}
	})

	// --- Update Game ---
	t.Run("Update Game", func(t *testing.T) {
		updateData := map[string]int{"day": 5}
		body, _ := json.Marshal(updateData)

		req := httptest.NewRequest(http.MethodPut, "/games/update?id="+gameID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.UpdateGame(w, req)

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

	// --- Delete Game ---
	t.Run("Delete Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/games/delete?id="+gameID.String(), nil)
		w := httptest.NewRecorder()

		handlers.DeleteGame(w, req)

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

	// --- Get Deleted Game (Expect Not Found) ---
	t.Run("Get Deleted Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/games/get?id="+gameID.String(), nil)
		w := httptest.NewRecorder()

		handlers.GetGame(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", res.StatusCode)
		}

		var response internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Success {
			t.Fatal("Expected failure, got success")
		}

		if response.Error != internal.ErrGameNotFound {
			t.Fatalf("Expected error %s, got %s", internal.ErrGameNotFound, response.Error)
		}
	})
}
