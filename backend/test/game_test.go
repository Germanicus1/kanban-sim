package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Germanicus1/kanban-sim/handlers"
	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var db *sql.DB

func setupEnv(t *testing.T) {
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
}

func setupDB(t *testing.T) {
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

	_, err = db.Exec("DELETE FROM games")
	if err != nil {
		t.Fatalf("Failed to clean games table: %v", err)
	}
}

func tearDownDB() {
	if db != nil {
		db.Close()
	}
}

func TestGameCRUD(t *testing.T) {
	setupDB(t)
	defer tearDownDB()

	var gameID uuid.UUID

	t.Run("Create Game", func(t *testing.T) {
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

	t.Run("Get Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/games/get?id="+gameID.String(), nil)
		w := httptest.NewRecorder()

		handlers.GetGame(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", res.StatusCode)
		}

		var game handlers.Game
		if err := json.NewDecoder(res.Body).Decode(&game); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if game.ID != gameID {
			t.Fatalf("Expected game ID %s, got %s", gameID, game.ID)
		}
	})

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

		// Verify the update
		req = httptest.NewRequest(http.MethodGet, "/games/get?id="+gameID.String(), nil)
		w = httptest.NewRecorder()
		handlers.GetGame(w, req)

		res = w.Result()
		defer res.Body.Close()

		var game handlers.Game
		if err := json.NewDecoder(res.Body).Decode(&game); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if game.Day != 5 {
			t.Fatalf("Expected day 5, got %d", game.Day)
		}
	})

	t.Run("Delete Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/games/delete?id="+gameID.String(), nil)
		w := httptest.NewRecorder()

		handlers.DeleteGame(w, req)

		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected status 204, got %d", res.StatusCode)
		}

		// Verify deletion
		req = httptest.NewRequest(http.MethodGet, "/games/get?id="+gameID.String(), nil)
		w = httptest.NewRecorder()
		handlers.GetGame(w, req)

		res = w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected status 404, got %d", res.StatusCode)
		}
	})
}
