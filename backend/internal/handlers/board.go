package handlers

import (
	"database/sql"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal/config"
	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/Germanicus1/kanban-sim/internal/response"
	"github.com/google/uuid"
)

// GetBoard returns all cards for a game, grouped by column.
func GetBoard(w http.ResponseWriter, r *http.Request) {
	// 1) Validate game_id
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	// 2) Load config so we can initialize empty slices for every column group
	cfg, err := config.LoadBoardConfig()
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"failed to load board config: "+err.Error())
		return
	}
	board := make(map[string][]models.Card, len(cfg.Columns)*2)
	for _, col := range cfg.Columns {
		board[col.Title] = nil
		for _, sub := range col.SubColumns {
			board[col.Title+" - "+sub.Title] = nil
		}
	}

	// 3) Query cards + human-readable column title
	rows, err := database.DB.Query(`
        SELECT
            c.id, c.game_id, c.column_id, c.title,
            CASE
              WHEN parent.id IS NULL THEN col.title
              ELSE parent.title || ' - ' || col.title
            END AS column_title,
            c.class_of_service, c.value_estimate,
            c.selected_day, c.deployed_day
        FROM cards c
        JOIN columns col        ON col.id = c.column_id
        LEFT JOIN columns parent ON parent.id = col.parent_id
        WHERE c.game_id = $1
        ORDER BY c.selected_day
    `, gameID)
	if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}
	defer rows.Close()

	// 4) Scan each card into models.Card, then load its efforts
	for rows.Next() {
		var (
			c        models.Card
			cs, ve   sql.NullString
			sel, dep sql.NullInt64
		)
		if err := rows.Scan(
			&c.ID,
			&c.GameID,
			&c.ColumnID,
			&c.Title,
			&c.ColumnTitle,
			&cs,
			&ve,
			&sel,
			&dep,
		); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError,
				response.ErrDatabaseError)
			return
		}

		// Null → zero‐value conversion
		if cs.Valid {
			c.ClassOfService = cs.String
		}
		if ve.Valid {
			c.ValueEstimate = ve.String
		}
		if sel.Valid {
			c.SelectedDay = int(sel.Int64)
		}
		if dep.Valid {
			c.DeployedDay = int(dep.Int64)
		}

		// 5) Load efforts for this card
		effortRows, err := database.DB.Query(`
            SELECT et.title, e.estimate, e.remaining, e.actual
              FROM efforts e
              JOIN effort_types et ON et.id = e.effort_type_id
             WHERE e.card_id = $1
             ORDER BY et.order_index
        `, c.ID)
		if err == nil {
			for effortRows.Next() {
				var (
					etTitle  string
					est      int
					rem, act sql.NullInt64
				)
				if err := effortRows.Scan(&etTitle, &est, &rem, &act); err != nil {
					break
				}
				e := models.Effort{
					EffortType: etTitle,
					Estimate:   est,
				}
				if rem.Valid {
					tmp := int(rem.Int64)
					e.Remaining = &tmp
				}
				if act.Valid {
					tmp := int(act.Int64)
					e.Actual = &tmp
				}
				c.Efforts = append(c.Efforts, e)
			}
			effortRows.Close()
		}

		board[c.ColumnTitle] = append(board[c.ColumnTitle], c)
	}

	// 6) Return the grouped map[string][]models.Card
	response.RespondWithData(w, board)
}
