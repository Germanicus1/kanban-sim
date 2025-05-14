package handlers_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Germanicus1/kanban-sim/internal/config"
	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/handlers"
	"github.com/google/uuid"
)

func TestCreateGame_CardsAndEffortsMatchConfig(t *testing.T) {
	// 1) clean only the tables we’re going to check
	SetupDB(t, "games")
	SetupDB(t, "cards")
	SetupDB(t, "effort_types")
	SetupDB(t, "efforts")
	defer TearDownDB()

	// 2) load your board config (with ConfigCard entries)
	cfg, err := config.LoadBoardConfig()
	if err != nil {
		t.Fatalf("LoadBoardConfig: %v", err)
	}

	// 3) exercise the handler
	req := httptest.NewRequest(http.MethodPost, "/games", nil)
	w := httptest.NewRecorder()
	handlers.CreateGame(w, req)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("CreateGame returned status %d", res.StatusCode)
	}

	// 4) extract the gameID from the JSON envelope
	var enveloped struct {
		Success bool `json:"success"`
		Data    struct {
			ID uuid.UUID `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(res.Body).Decode(&enveloped); err != nil {
		t.Fatalf("decode CreateGame response: %v", err)
	}
	gameID := enveloped.Data.ID

	// 5) sanity-check that effort_types exist
	wantET := len(cfg.EffortTypes)
	gotET := mustCount(t, `SELECT COUNT(*) FROM effort_types WHERE game_id=$1`, gameID)
	if gotET != wantET {
		t.Fatalf("effort_types: got %d rows, want %d", gotET, wantET)
	}

	// 6) for each card in the config, verify its DB row + efforts
	for _, want := range cfg.Cards {
		t.Run("card "+want.Title, func(t *testing.T) {
			// fetch the card row
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
				t.Fatalf("card %q scan failed: %v", want.Title, err)
			}

			// assert class_of_service
			if want.ClassOfService != nil {
				if !cls.Valid || cls.String != *want.ClassOfService {
					t.Errorf("classOfService = %v; want %v", cls.String, *want.ClassOfService)
				}
			} else if cls.Valid {
				t.Errorf("expected no classOfService, got %q", cls.String)
			}

			// assert value_estimate
			if want.ValueEstimate != nil {
				if !val.Valid || val.String != *want.ValueEstimate {
					t.Errorf("valueEstimate = %v; want %v", val.String, *want.ValueEstimate)
				}
			} else if val.Valid {
				t.Errorf("expected no valueEstimate, got %q", val.String)
			}

			// assert selected_day
			if want.SelectedDay != nil {
				if !sel.Valid || int(sel.Int64) != *want.SelectedDay {
					t.Errorf("selectedDay = %v; want %v", sel.Int64, *want.SelectedDay)
				}
			}

			// assert deployed_day
			if want.DeployedDay != nil {
				if !dep.Valid || int(dep.Int64) != *want.DeployedDay {
					t.Errorf("deployedDay = %v; want %v", dep.Int64, *want.DeployedDay)
				}
			}

			// now fetch and assert efforts
			rows, err := database.DB.Query(`
                SELECT et.title, e.estimate
                  FROM efforts e
                  JOIN effort_types et ON et.id=e.effort_type_id
                 WHERE e.card_id=$1
              ORDER BY et.order_index
            `, cardID)
			if err != nil {
				t.Fatalf("query efforts for card %q failed: %v", want.Title, err)
			}
			defer rows.Close()

			i := 0
			for rows.Next() {
				var gotType string
				var gotEst int
				if err := rows.Scan(&gotType, &gotEst); err != nil {
					t.Fatalf("scan effort[%d] for %q failed: %v", i, want.Title, err)
				}
				if i >= len(want.Efforts) {
					t.Fatalf("unexpected extra effort %q/%d", gotType, gotEst)
				}
				exp := want.Efforts[i]
				if gotType != exp.EffortType || gotEst != exp.Estimate {
					t.Errorf("effort[%d] = (%q,%d); want (%q,%d)",
						i, gotType, gotEst, exp.EffortType, exp.Estimate)
				}
				i++
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("rows error for %q: %v", want.Title, err)
			}
			if i != len(want.Efforts) {
				t.Errorf("for card %q, expected %d efforts, got %d", want.Title, len(want.Efforts), i)
			}
		})
	}
}
