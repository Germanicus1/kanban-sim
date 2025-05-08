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

func setupCardTestDB(t *testing.T) {
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

	_, err = db.Exec("DELETE FROM cards")
	if err != nil {
		t.Fatalf("Failed to clean cards table: %v", err)
	}

	_, err = db.Exec("DELETE FROM games")
	if err != nil {
		t.Fatalf("Failed to clean games table: %v", err)
	}
}

func TestCardCRUD(t *testing.T) {
	setupCardTestDB(t)
	defer tearDownDB()

	// First, create a game to associate with the card
	var gameID uuid.UUID
	t.Run("Create Game for Card", func(t *testing.T) {
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

	var cardID uuid.UUID

	// --- Create Card ---
	t.Run("Create Card", func(t *testing.T) {
		cardData := map[string]interface{}{
			"game_id":     gameID.String(),
			"title":       "New Card",
			"card_column": "options",
		}
		body, _ := json.Marshal(cardData)

		req := httptest.NewRequest(http.MethodPost, "/cards", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handlers.CreateCard(w, req)

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

		cardID, _ = uuid.Parse(idStr)
		if cardID == uuid.Nil {
			t.Fatal("Expected valid UUID, got Nil")
		}
	})

	// --- Get Card ---
	t.Run("Get Card", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		w := httptest.NewRecorder()

		handlers.GetCard(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		var card handlers.Card
		if err := json.NewDecoder(res.Body).Decode(&card); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if card.ID != cardID {
			t.Fatalf("Expected card ID %s, got %s", cardID, card.ID)
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

		// Verify the update
		req = httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		w = httptest.NewRecorder()
		handlers.GetCard(w, req)

		res = w.Result()
		defer res.Body.Close()

		var card handlers.Card
		if err := json.NewDecoder(res.Body).Decode(&card); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if card.Title != "Updated Card" {
			t.Fatalf("Expected title 'Updated Card', got '%s'", card.Title)
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

		// Verify deletion
		req = httptest.NewRequest(http.MethodGet, "/cards/get?id="+cardID.String(), nil)
		w = httptest.NewRecorder()
		handlers.GetCard(w, req)

		res = w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", res.StatusCode)
		}
	})
}
