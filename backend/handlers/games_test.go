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

func TestGameCRUD(t *testing.T) {
	setupDB(t, "games")
	setupDB(t, "cards")
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

	// --- Verify Columns ---
	t.Run("Verify Columns", func(t *testing.T) {
		rows, err := internal.DB.Query(`SELECT title, column_order FROM cards WHERE game_id = $1 ORDER BY column_order`, gameID)
		if err != nil {
			t.Fatalf("Failed to query columns: %v", err)
		}
		defer rows.Close()

		expectedColumns := []string{"Options", "Selected", "Analysis - In Progress", "Analysis - Done", "Development - In Progress", "Development - Done", "Test", "Ready to Deploy", "Deployed"}

		index := 0
		for rows.Next() {
			var title string
			var order int
			if err := rows.Scan(&title, &order); err != nil {
				t.Fatalf("Failed to scan column: %v", err)
			}

			if title != expectedColumns[index] {
				t.Errorf("Expected column %s at index %d, got %s", expectedColumns[index], index, title)
			}
			index++
		}
	})

	// --- Verify Cards ---
	t.Run("Verify Cards", func(t *testing.T) {
		rows, err := internal.DB.Query(`SELECT id, card_column FROM cards WHERE game_id = $1`, gameID)
		if err != nil {
			t.Fatalf("Failed to query cards: %v", err)
		}
		defer rows.Close()

		count := 0
		for rows.Next() {
			count++
		}

		if count == 0 {
			t.Error("Expected cards to be inserted, found none")
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

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected status 204, got %d", res.StatusCode)
		}
	})

	// --- Delete Game ---
	t.Run("Delete Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/games/delete?id="+gameID.String(), nil)
		w := httptest.NewRecorder()

		handlers.DeleteGame(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected status 204, got %d", res.StatusCode)
		}
	})
}
