package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Germanicus1/kanban-sim/internal/config"
	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/Germanicus1/kanban-sim/internal/response"
	"github.com/google/uuid"
)

// CreateGame creates a new game and seeds effort_types, columns, cards, and efforts.
func CreateGame(w http.ResponseWriter, r *http.Request) {
	// 0) load config
	cfg, err := config.LoadBoardConfig()
	_ = cfg
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"configuration load failed: "+err.Error())
		return
	}

	// 1) begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"failed to begin transaction")
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			tx.Rollback() // rollback only if there was an error
		}
	}()

	// 2) insert game
	gameID := uuid.New()
	if _, err := tx.Exec(
		`INSERT INTO games (id, created_at, day) VALUES ($1, NOW(), 1)`,
		gameID,
	); err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"failed to insert game: "+err.Error())
		return
	}

	// 3) seed effort_types
	effortTypeIDs := make(map[string]uuid.UUID, len(cfg.EffortTypes))
	for idx, et := range cfg.EffortTypes {
		etID := uuid.New()
		if _, err := tx.Exec(
			`INSERT INTO effort_types (id, game_id, title, order_index)
			VALUES ($1,$2,$3,$4)`,
			etID, gameID, et.Title, idx,
		); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError,
				"failed to insert effort type: "+err.Error())
			return
		}
		effortTypeIDs[et.Title] = etID
	}

	// 4) seed columns & subcolumns
	columnIDs := make(map[string]uuid.UUID, len(cfg.Columns)*2)
	for _, col := range cfg.Columns {
		mainID := uuid.New()
		if _, err := tx.Exec(
			`INSERT INTO columns (id, game_id, title, parent_id, order_index)
					VALUES ($1,$2,$3,NULL,$4)`,
			mainID, gameID, col.Title, col.OrderIndex,
		); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "failed to insert column: "+err.Error())
			return
		}
		columnIDs[col.Title] = mainID

		for _, sub := range col.SubColumns {
			subID := uuid.New()
			if _, err := tx.Exec(
				`INSERT INTO columns (id, game_id, title, parent_id, order_index)
				VALUES ($1,$2,$3,$4,$5)`,
				subID, gameID, sub.Title, mainID, sub.OrderIndex,
			); err != nil {
				msg := fmt.Sprintf("failed to insert subcolumn %q under %q: %v", sub.Title, col.Title, err)
				response.RespondWithError(w, http.StatusInternalServerError, msg)
				return
			}
			columnIDs[col.Title+" - "+sub.Title] = subID
		}
	}

	// REF  Seeding effort_types, columns, cards, and efforts

	// 5) seed cards & their efforts

	for _, c := range cfg.Cards {
		cardID := uuid.New()
		colID, ok := columnIDs[c.ColumnTitle]
		if !ok {
			response.RespondWithError(w, http.StatusInternalServerError, "unknown column "+colID.String())
			return
		}
		// _ = cardID
		// _ = colID
		if _, err := tx.Exec(
			`INSERT INTO cards
	   					(id, game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day)
	   					VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
			cardID, gameID, colID,
			c.Title, c.ClassOfService, c.ValueEstimate,
			c.SelectedDay, c.DeployedDay,
		); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "failed to insert card: "+err.Error())
			return
		}

		for _, e := range c.Efforts {
			etID, ok := effortTypeIDs[e.EffortType]
			if !ok {
				response.RespondWithError(w, http.StatusInternalServerError, "unknown effort type "+e.EffortType)
				return
			}
			if _, err := tx.Exec(
				`INSERT INTO efforts (id, card_id, effort_type_id, estimate, remaining, actual)
							VALUES($1,$2,$3,$4,$4,0)`,
				uuid.New(), cardID, etID, e.Estimate,
			); err != nil {
				response.RespondWithError(w, http.StatusInternalServerError, "failed to insert effort: "+err.Error())
				return
			}
		}
	}

	// 6) commit
	if err := tx.Commit(); err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
		return
	}

	// 7) respond
	response.RespondWithData(w, map[string]uuid.UUID{"id": gameID})
}

// GetGame retrieves a game by ID.
func GetGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	const q = `SELECT id, created_at, day FROM games WHERE id = $1`

	var g models.Game

	err = database.DB.QueryRow(q, id).Scan(&g.ID, &g.CreatedAt, &g.Day)
	if err == sql.ErrNoRows {
		response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		return
	} else if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}

	response.RespondWithData(w, g)
}

// UpdateGame updates the day of a game.
func UpdateGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	var payload struct {
		Day int `json:"day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrValidationFailed)
		return
	}

	const q = `UPDATE games SET day = $1 WHERE id = $2`
	res, err := database.DB.Exec(q, payload.Day, id)
	if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		return
	}

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// DeleteGame deletes a game by ID.
func DeleteGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	const q = `DELETE FROM games WHERE id = $1`
	_, err = database.DB.Exec(q, id)
	if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}
	// affected, _ := res.RowsAffected()
	// if affected == 0 {
	// 	response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
	// 	return
	// }

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// GetEvents returns the audit log for a game.
func GetEvents(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.URL.Query().Get("game_id"))
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	rows, err := database.DB.Query(
		`SELECT id, card_id, event_type, payload, created_at 
           FROM game_events 
          WHERE game_id=$1 
       ORDER BY created_at`,
		gameID,
	)
	if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}
	defer rows.Close()

	var events []any
	for rows.Next() {
		var (
			id        uuid.UUID
			cardID    uuid.UUID
			typ       string
			raw       []byte
			createdAt time.Time
		)
		rows.Scan(&id, &cardID, &typ, &raw, &createdAt)
		events = append(events, map[string]any{
			"id":         id,
			"card_id":    cardID,
			"event_type": typ,
			"payload":    json.RawMessage(raw),
			"created_at": createdAt,
		})
	}
	response.RespondWithData(w, events)
}

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
