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

	// Create Card
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

	// Move Card
	t.Run("Move Card", func(t *testing.T) {
		// fetch current column to ensure correct 'from'
		reqGet := httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		wGet := httptest.NewRecorder()
		handlers.GetCard(wGet, reqGet)
		resGet := wGet.Result()
		defer resGet.Body.Close()
		var respGet internal.APIResponse
		json.NewDecoder(resGet.Body).Decode(&respGet)
		mGet := respGet.Data.(map[string]interface{})
		from := mGet["card_column"].(string)

		payload := map[string]interface{}{
			"from_column": from,
			"to_column":   "selected",
			"day":         3,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/cards/move?id="+cardID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handlers.MoveCard(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected 204, got %d", res.StatusCode)
		}

		// Verify column changed
		req2 := httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		w2 := httptest.NewRecorder()
		handlers.GetCard(w2, req2)
		res2 := w2.Result()
		defer res2.Body.Close()

		if res2.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res2.StatusCode)
		}
		var resp2 internal.APIResponse
		json.NewDecoder(res2.Body).Decode(&resp2)
		m2 := resp2.Data.(map[string]interface{})
		if m2["card_column"].(string) != "selected" {
			t.Fatalf("Expected column %q, got %q", "selected", m2["card_column"].(string))
		}
	})

	// Update Card
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

	t.Run("Move Card – invalid from", func(t *testing.T) {
		bad := map[string]interface{}{
			"from_column": "not-actually-there",
			"to_column":   "selected",
			"day":         3,
		}
		b, _ := json.Marshal(bad)
		req := httptest.NewRequest(http.MethodPost, "/cards/move?id="+cardID.String(), bytes.NewBuffer(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handlers.MoveCard(w, req)

		if w.Result().StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d", w.Result().StatusCode)
		}
		var resp internal.APIResponse
		json.NewDecoder(w.Result().Body).Decode(&resp)
		if resp.Error != "INVALID_MOVE_FROM" {
			t.Fatalf("Expected INVALID_MOVE_FROM, got %q", resp.Error)
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
