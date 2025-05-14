package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/internal/config"
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

	cfg, err := config.LoadBoardConfig()
	if err != nil {
		t.Fatalf("failed to load board config: %v", err)
	}

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

	t.Run("Verify Efforts", func(t *testing.T) {
		// 2) effort_types
		if got := mustCount(t, `SELECT COUNT(*) FROM effort_types WHERE game_id=$1`, gameID); got != len(cfg.EffortTypes) {
			t.Errorf("effort_types: expected %d, got %d", len(cfg.EffortTypes), got)
		}
	})

	// --- Verify Columns ---
	t.Run("Verify Columns", func(t *testing.T) {
		rows, err := database.DB.Query(`
        SELECT CASE
            WHEN c.parent_id IS NULL THEN c.title
            ELSE p.title || ' - ' || c.title
        END AS full_title
        FROM columns c
        LEFT JOIN columns p ON p.id = c.parent_id
        WHERE c.game_id = $1
        ORDER BY
            COALESCE(p.order_index, c.order_index),  -- parent position
            c.order_index                            -- subposition
    `, gameID)
		if err != nil {
			t.Fatalf("Failed to query columns: %v", err)
		}
		defer rows.Close()

		expected := []string{
			"Options",
			"Selected",
			"Analysis - In Progress",
			"Analysis - Done",
			"Analysis",
			"Development - In Progress",
			"Development - Done",
			"Development",
			"Test",
			"Ready to Deploy",
			"Deployed",
		}

		var got []string
		for rows.Next() {
			var title string
			if err := rows.Scan(&title); err != nil {
				t.Fatalf("Scan failed: %v", err)
			}
			got = append(got, title)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("Rows error: %v", err)
		}

		if len(got) != len(expected) {
			t.Fatalf("wrong number of columns: got %d, want %d", len(got), len(expected))
		}

		log.Printf("Columns: %v\n", got)

		for i := range expected {
			if got[i] != expected[i] {
				t.Errorf("column[%d]: got %q, want %q", i, got[i], expected[i])
			}
		}
	})

	// --- Verify Cards ---
	t.Run("Verify Cards Count", func(t *testing.T) {
		// t.Skip("skipping card verification")
		rows, err := database.DB.Query(`SELECT id, column_id FROM cards WHERE game_id = $1`, gameID)
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

	for _, want := range cfg.Cards {
		t.Run("Verify Cards "+want.Title, func(t *testing.T) {
			// fetch the card
			row := database.DB.QueryRow(`
                SELECT id, column_id, class_of_service, value_estimate, selected_day, deployed_day
                  FROM cards
                 WHERE game_id=$1 AND title=$2
            `, gameID, want.Title)

			var (
				cardID uuid.UUID
				colID  uuid.UUID
				cls    sql.NullString
				val    sql.NullString
				sel    sql.NullInt64
				dep    sql.NullInt64
			)
			if err := row.Scan(&cardID, &colID, &cls, &val, &sel, &dep); err != nil {
				t.Fatalf("card %q not found or scan failed: %v", want.Title, err)
			}

			// assert class_of_service
			if want.ClassOfService != nil {
				if !cls.Valid || cls.String != *want.ClassOfService {
					t.Errorf("classOfService: got %v; want %v", cls.String, *want.ClassOfService)
				}
			} else if cls.Valid {
				t.Errorf("classOfService unexpectedly %q", cls.String)
			}

			// assert value_estimate
			if want.ValueEstimate != nil {
				if !val.Valid || val.String != *want.ValueEstimate {
					t.Errorf("valueEstimate: got %v; want %v", val.String, *want.ValueEstimate)
				}
			} else if val.Valid {
				t.Errorf("valueEstimate unexpectedly %q", val.String)
			}

			// assert selected_day
			if want.SelectedDay != nil {
				if !sel.Valid || int(sel.Int64) != *want.SelectedDay {
					t.Errorf("selectedDay: got %v; want %v", sel.Int64, *want.SelectedDay)
				}
			}

			// assert deployed_day
			if want.DeployedDay != nil {
				if !dep.Valid || int(dep.Int64) != *want.DeployedDay {
					t.Errorf("deployedDay: got %v; want %v", dep.Int64, *want.DeployedDay)
				}
			}

			// now check efforts
			rows, err := database.DB.Query(`
                SELECT et.title, e.estimate
                  FROM efforts e
                  JOIN effort_types et ON e.effort_type_id=et.id
                 WHERE e.card_id=$1
              ORDER BY et.order_index
            `, cardID)
			if err != nil {
				t.Fatalf("query efforts failed: %v", err)
			}
			defer rows.Close()

			i := 0
			for rows.Next() {
				var gotType string
				var gotEst int
				if err := rows.Scan(&gotType, &gotEst); err != nil {
					t.Fatalf("scan effort: %v", err)
				}
				if i >= len(want.Efforts) {
					t.Fatalf("extra effort %q/%d", gotType, gotEst)
				}
				wantEff := want.Efforts[i]
				if gotType != wantEff.EffortType || gotEst != wantEff.Estimate {
					t.Errorf("effort[%d]: got (%q,%d); want (%q,%d)",
						i, gotType, gotEst, wantEff.EffortType, wantEff.Estimate)
				}
				i++
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("rows err: %v", err)
			}
			if i != len(want.Efforts) {
				t.Errorf("expected %d efforts, got %d", len(want.Efforts), i)
			}
		})
	}

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
