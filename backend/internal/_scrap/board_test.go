package handlers_test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/Germanicus1/kanban-sim/internal/handlers"
	"github.com/google/uuid"
)

func TestGetBoard(t *testing.T) {
	// Clean and prepare
	setupDB(t, "cards")
	setupDB(t, "games")
	defer tearDownDB()

	var gameID uuid.UUID

	// 1) Create a game
	t.Run("Create Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/games", nil)
		w := httptest.NewRecorder()
		handlers.CreateGame(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			var errResp internal.APIResponse
			_ = json.NewDecoder(res.Body).Decode(&errResp)
			t.Fatalf("CreateGame: expected 200, got %d: %s", res.StatusCode, errResp.Error)
		}

		var resp internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("CreateGame decode: %v", err)
		}
		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("CreateGame: expected response.Data to be a map")
		}
		idStr, ok := data["id"].(string)
		if !ok || idStr == "" {
			t.Fatal("CreateGame: expected a valid 'id'")
		}
		var err error
		gameID, err = uuid.Parse(idStr)
		if err != nil {
			t.Fatalf("CreateGame parse ID: %v", err)
		}
	})

	// 2) Invoke GetBoard
	t.Run("GetBoard", func(t *testing.T) {

		log.Printf("gameID: %s", gameID.String())
		req := httptest.NewRequest(http.MethodGet, "/games/board?game_id="+gameID.String(), nil)
		w := httptest.NewRecorder()
		handlers.GetBoard(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			var errResp internal.APIResponse
			if decodeErr := json.NewDecoder(res.Body).Decode(&errResp); decodeErr != nil {
				t.Fatalf("GetBoard: expected 200, got %d and failed to decode error: %v", res.StatusCode, decodeErr)
			}
			t.Fatalf("GetBoard: expected 200, got %d: %s", res.StatusCode, errResp.Error)
		}

		var resp internal.APIResponse
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("GetBoard decode: %v", err)
		}
		if !resp.Success {
			t.Fatalf("GetBoard: expected success, got error %q", resp.Error)
		}
	})
}
