package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/backend/internal/config"
	"github.com/Germanicus1/kanban-sim/backend/internal/handlers"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

func TestGetBoard_ValidationErrors(t *testing.T) {
	// Missing game_id
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/games/board", nil)
	handlers.GetBoard(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("no game_id → status %d; want %d", rec.Code, http.StatusBadRequest)
	}
	var env1 response.APIResponse[any]
	if err := json.NewDecoder(rec.Body).Decode(&env1); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if env1.Error != response.ErrInvalidGameID {
		t.Errorf("no game_id → error %q; want %q", env1.Error, response.ErrInvalidGameID)
	}

	// Invalid UUID
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/games/board?game_id=not-a-uuid", nil)
	handlers.GetBoard(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bad UUID → status %d; want %d", rec.Code, http.StatusBadRequest)
	}
	var env2 response.APIResponse[any]
	if err := json.NewDecoder(rec.Body).Decode(&env2); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if env2.Error != response.ErrInvalidGameID {
		t.Errorf("bad UUID → error %q; want %q", env2.Error, response.ErrInvalidGameID)
	}
}

func TestGetBoard_Success(t *testing.T) {
	// 1) Reset and seed via CreateGame
	SetupDB(t, "games")
	SetupDB(t, "columns")
	SetupDB(t, "cards")
	SetupDB(t, "effort_types")
	SetupDB(t, "efforts")
	defer TearDownDB()

	// Create a new game which seeds everything
	createRec := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/games", nil)
	handlers.CreateGame(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("CreateGame → status %d, want %d", createRec.Code, http.StatusOK)
	}
	var createEnv struct {
		Success bool `json:"success"`
		Data    struct {
			ID uuid.UUID `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(createRec.Body).Decode(&createEnv); err != nil {
		t.Fatalf("decode CreateGame: %v", err)
	}
	gameID := createEnv.Data.ID

	// Load board config for expected structure
	cfg, err := config.LoadBoardConfig()
	if err != nil {
		t.Fatalf("LoadBoardConfig(): %v", err)
	}

	// 2) Call GetBoard
	boardRec := httptest.NewRecorder()
	boardReq := httptest.NewRequest(
		http.MethodGet,
		"/games/board?game_id="+gameID.String(),
		nil,
	)
	handlers.GetBoard(boardRec, boardReq)
	if boardRec.Code != http.StatusOK {
		t.Fatalf("GetBoard → status %d, want %d", boardRec.Code, http.StatusOK)
	}

	// 3) Decode into envelope of map[string][]handlers.Card
	var boardEnv response.APIResponse[map[string][]handlers.Card]
	if err := json.NewDecoder(boardRec.Body).Decode(&boardEnv); err != nil {
		t.Fatalf("decode GetBoard: %v", err)
	}
	if !boardEnv.Success {
		t.Fatalf("GetBoard error: %s", boardEnv.Error)
	}
	board := boardEnv.Data

	// 4) Each configured column must appear (even if empty)
	for _, col := range cfg.Columns {
		if _, exists := board[col.Title]; !exists {
			t.Errorf("missing column group %q", col.Title)
		}
	}

	// 5) Each configured card must live under its ColumnTitle
	for _, want := range cfg.Cards {
		group, ok := board[want.ColumnTitle]
		if !ok {
			t.Errorf("column %q not found for card %q", want.ColumnTitle, want.Title)
			continue
		}
		found := false
		for _, c := range group {
			if c.Title == want.Title {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("card %q missing under column %q", want.Title, want.ColumnTitle)
		}
	}
}
