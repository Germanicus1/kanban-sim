package handlers_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/handlers"
	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/Germanicus1/kanban-sim/internal/response"
	"github.com/google/uuid"
)

type gamePayload struct {
	ID uuid.UUID `json:"id"`
}

// mustCount runs a COUNT(*) query and fatals the test if anything goes wrong.
func mustCount(t *testing.T, query string, args ...interface{}) int {
	t.Helper() // marks this function as a test helper
	var n int
	if err := database.DB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("query %q failed: %v", query, err)
	}
	return n
}

func loadConfig(t *testing.T) *models.Board {
	t.Helper()
	path := filepath.Join("..", "config", "board_config.json")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("cannot open config: %v", err)
	}
	defer f.Close()

	var cfg models.Board
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		t.Fatalf("cannot decode config: %v", err)
	}
	return &cfg
}
func TestGameCRUD(t *testing.T) {
	SetupDB(t, "games")
	SetupDB(t, "cards")
	defer TearDownDB()

	var gameID uuid.UUID

	// --- Create Game ---
	t.Run("Create Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/games", nil)
		w := httptest.NewRecorder()
		handlers.CreateGame(w, req)
		res := w.Result()
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.StatusCode)
		}

		var resp response.APIResponse[gamePayload]
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("decode resp: %v", err)
		}

		// assert success
		if !resp.Success {
			t.Fatalf("expected success=true, got error: %s", resp.Error)
		}

		// now resp.Data is an idPayload, no casting needed
		if resp.Data.ID == uuid.Nil {
			t.Fatal("expected non-nil UUID in resp.Data.ID")
		}
		gameID = resp.Data.ID
	})

	t.Run("Verify Seeding", func(t *testing.T) {
		cfg := loadConfig(t)
		log.Printf("Efforts: %v\n", len(cfg.EffortTypes))
		log.Println(gameID)

		_ = cfg // just to silence the compiler since we are not using cfg yet.

		// 1) games
		if got := mustCount(t, `SELECT COUNT(*) FROM games WHERE id=$1`, gameID); got != 1 {
			t.Errorf("games: expected 1, got %d", got)
		}

		// // 2) effort_types
		// if got := mustCount(t, `SELECT COUNT(*) FROM effort_types WHERE game_id=$1`, gameID); got != len(cfg.EffortTypes) {
		// 	t.Errorf("effort_types: expected %d, got %d", len(cfg.EffortTypes), got)
		// }

		/* 		// 3) columns + subcolumns
		   		totalCols := 0
		   		for _, c := range cfg.Columns {
		   			totalCols++
		   			totalCols += len(c.SubColumns)
		   		}
		   		if got := mustCount(t, `SELECT COUNT(*) FROM columns WHERE game_id=$1`, gameID); got != totalCols {
		   			t.Errorf("columns: expected %d, got %d", totalCols, got)
		   		}
		   		// 4) cards
		   		if got := mustCount(t, `SELECT COUNT(*) FROM cards WHERE game_id=$1`, gameID); got != len(cfg.Cards) {
		   			t.Errorf("cards: expected %d, got %d", len(cfg.Cards), got)
		   		}
		   		// 5) efforts
		   		sumEff := 0
		   		for _, c := range cfg.Cards {
		   			sumEff += len(c.Efforts)
		   		}
		   		if got := mustCount(t, `
		   		    SELECT COUNT(*) FROM efforts
		   		     WHERE card_id IN (SELECT id FROM cards WHERE game_id=$1)
		   		`, gameID); got != sumEff {
		   			t.Errorf("efforts: expected %d, got %d", sumEff, got)
		   		} */
	})

	// --- Verify Columns ---
	t.Run("Verify Columns", func(t *testing.T) {
		t.Skip("columns seeding not implemented yet")
		rows, err := database.DB.Query(`SELECT title, column_order FROM cards WHERE game_id = $1 ORDER BY column_order`, gameID)
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
		t.Skip("cards seeding not implemented yet")
		rows, err := database.DB.Query(`SELECT id, card_column FROM cards WHERE game_id = $1`, gameID)
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

	// --- Get Game ---
	t.Run("Get Game", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/games/get?id="+gameID.String(), nil)
		w := httptest.NewRecorder()
		handlers.GetGame(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("Expected 200, got %d", res.StatusCode)
		}
		var resp response.APIResponse[gamePayload]
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("decode resp: %v", err)
		}
		if !resp.Success {
			t.Fatalf("expected success=true, got error: %s", resp.Error)
		}
		if resp.Data.ID != gameID {
			t.Fatalf("expected gameID %s, got %s", gameID.String(), resp.Data.ID.String())
		}
		if resp.Data.ID == uuid.Nil {
			t.Fatal("expected non-nil UUID in resp.Data.ID")
		}
		if resp.Data.ID != gameID {
			t.Fatalf("expected gameID %s, got %s", gameID.String(), resp.Data.ID.String())
		}
	})
	// --- Get Game Not Found ---
	t.Run("Get Game Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/games/get?id="+uuid.New().String(), nil)
		w := httptest.NewRecorder()
		handlers.GetGame(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusNotFound {
			t.Fatalf("Expected 404, got %d", res.StatusCode)
		}
		var resp response.APIResponse[gamePayload]
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("decode resp: %v", err)
		}
		if resp.Success {
			t.Fatalf("expected success=false, got error: %s", resp.Error)
		}
		if resp.Data.ID != uuid.Nil {
			t.Fatalf("expected empty gameID, got %s", resp.Data.ID.String())
		}
		if resp.Error != response.ErrGameNotFound {
			t.Fatalf("expected error %s, got %s", response.ErrGameNotFound, resp.Error)
		}
	})
	// --- Get Game Bad Request ---
	t.Run("Get Game Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/games/get?id=invalid-uuid", nil)
		w := httptest.NewRecorder()
		handlers.GetGame(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d", res.StatusCode)
		}
		var resp response.APIResponse[gamePayload]
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("decode resp: %v", err)
		}
		if resp.Success {
			t.Fatalf("expected success=false, got error: %s", resp.Error)
		}
		if resp.Data.ID != uuid.Nil {
			t.Fatalf("expected empty gameID, got %s", resp.Data.ID.String())
		}
		if resp.Error != response.ErrInvalidGameID {
			t.Fatalf("expected error %s, got %s", response.ErrInvalidGameID, resp.Error)
		}
	})
	// --- Get Game Bad Request No ID ---
	t.Run("Get Game Bad Request No ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/games/get", nil)
		w := httptest.NewRecorder()
		handlers.GetGame(w, req)
		res := w.Result()
		defer res.Body.Close()
		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d", res.StatusCode)
		}
		var resp response.APIResponse[gamePayload]
		if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
			t.Fatalf("decode resp: %v", err)
		}
		if resp.Success {
			t.Fatalf("expected success=false, got error: %s", resp.Error)
		}
		if resp.Data.ID != uuid.Nil {
			t.Fatalf("expected empty gameID, got %s", resp.Data.ID.String())
		}
		if resp.Error != response.ErrInvalidGameID {
			t.Fatalf("expected error %s, got %s", response.ErrInvalidGameID, resp.Error)
		}
	})

	// --- Update Game ---
	t.Run("Update Game", func(t *testing.T) {
		// t.Skip("Update Game not implemented yet")

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
		t.Skip("Delete Game not implemented yet")

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

// TODO: Continue with the seeding of events_types
