package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/handlers"
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

func TestGameCRUD(t *testing.T) {
	SetupDB(t, "games")
	SetupDB(t, "cards")
	SetupDB(t, "efforts")
	SetupDB(t, "columns")
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

// TODO: Continue with the seeding of events_types
