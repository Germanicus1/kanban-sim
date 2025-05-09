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
