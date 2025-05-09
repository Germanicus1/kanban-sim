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

func TestCardCRUD(t *testing.T) {
	setupDB(t, "cards")
	setupDB(t, "games")
	defer tearDownDB()

	var gameID uuid.UUID
	var cardID uuid.UUID

	// --- Create Game (Required for Card Association) ---
	t.Run("Create Game for Card", func(t *testing.T) {
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

		data, ok := response.Data.(map[string]any)
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

	// --- Create Card ---
	t.Run("Create Card", func(t *testing.T) {
		cardData := map[string]interface{}{
			"game_id":     gameID.String(),
			"title":       "Test Card",
			"card_column": "options",
		}
		body, _ := json.Marshal(cardData)

		req := httptest.NewRequest(http.MethodPost, "/cards", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreateCard(w, req)

		res := w.Result()
		defer res.Body.Close()

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

		cardID, _ = uuid.Parse(idStr)

		// Log the cardID for debugging
		// t.Logf("Created Card ID: %s", cardID)
	})

	// --- Get Card ---
	t.Run("Get Card", func(t *testing.T) {
		if cardID == uuid.Nil {
			t.Fatal("Card ID is nil. CreateCard test may have failed.")
		}

		// t.Logf("Get Card ID: %s", cardID)

		req := httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		w := httptest.NewRecorder()

		handlers.GetCard(w, req)

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
	})

	// --- Update Card ---
	t.Run("Update Card", func(t *testing.T) {
		updateData := map[string]string{
			"title":       "Updated Card",
			"card_column": "selected",
		}
		body, _ := json.Marshal(updateData)

		req := httptest.NewRequest(http.MethodPut, "/cards/update?id="+cardID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.UpdateCard(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected status 204, got %d", res.StatusCode)
		}
	})

	// --- Delete Card ---
	t.Run("Delete Card", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/cards/delete?id="+cardID.String(), nil)
		w := httptest.NewRecorder()

		handlers.DeleteCard(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected status 204, got %d", res.StatusCode)
		}
	})

	// --- Get Deleted Card (Expect Not Found) ---
	t.Run("Get Deleted Card", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		w := httptest.NewRecorder()

		handlers.GetCard(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", res.StatusCode)
		}

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", res.StatusCode)
		}
		var response internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Error != internal.ErrCardNotFound {
			t.Fatalf("Expected error %s, got %s", internal.ErrCardNotFound, response.Error)
		}
	})
}
