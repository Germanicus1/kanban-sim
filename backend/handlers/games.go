package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Germanicus1/kanban-sim/config"
	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
)

type Game struct {
	ID        uuid.UUID       `json:"id"`
	CreatedAt string          `json:"created_at"`
	Day       int             `json:"day"`
	Columns   json.RawMessage `json:"columns"`
}

const configPath = "config/board_config.json"

// loadBoardConfig loads the board configuration.
func loadBoardConfig() (*config.BoardConfig, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var cfg config.BoardConfig
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config JSON: %w", err)
	}

	return &cfg, nil
}

// CreateGame creates a new game and seeds it with the default configuration.
func CreateGame(w http.ResponseWriter, r *http.Request) {
	boardConfig, err := loadBoardConfig()
	if err != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, "Configuration not loaded: "+err.Error())
		return
	}

	gameID := uuid.New()
	createdAt := time.Now().Format(time.RFC3339)
	day := 1

	const insertGameQuery = `
		INSERT INTO games (id, created_at, day)
		VALUES ($1, $2, $3)
	`
	_, err = internal.DB.Exec(insertGameQuery, gameID, createdAt, day)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}

	// Insert Columns
	for index, column := range boardConfig.Columns {
		const insertColumnQuery = `
			INSERT INTO cards (
				id, game_id, title, card_column, column_order
			) VALUES ($1, $2, $3, $4, $5)
		`

		_, err = internal.DB.Exec(
			insertColumnQuery,
			uuid.New(),
			gameID,
			column.Name,
			column.ID,
			index,
		)

		if err != nil {
			status, code := internal.MapPostgresError(err)
			internal.RespondWithError(w, status, code)
			return
		}
	}

	// Insert Cards
	for _, card := range boardConfig.InitialCards {
		const insertCardQuery = `
			INSERT INTO cards (
				id, game_id, title, card_column, class_of_service, value_estimate,
				effort_analysis, effort_development, effort_test, selected_day, deployed_day
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`

		_, err = internal.DB.Exec(
			insertCardQuery,
			uuid.New(),
			gameID,
			card.ID,
			card.ColumnID,
			card.ClassOfService,
			card.ValueEstimate,
			card.Effort.Analysis,
			card.Effort.Development,
			card.Effort.Test,
			card.SelectedDay,
			card.DeployedDay,
		)

		if err != nil {
			status, code := internal.MapPostgresError(err)
			internal.RespondWithError(w, status, code)
			return
		}
	}

	internal.RespondWithData(w, map[string]interface{}{
		"id":        gameID,
		"createdAt": createdAt,
		"day":       day,
	})
}

// GetGame retrieves a game by ID.
func GetGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	const q = `SELECT id, created_at, day, columns FROM games WHERE id = $1`
	var g Game
	err = internal.DB.QueryRow(q, id).Scan(&g.ID, &g.CreatedAt, &g.Day, &g.Columns)
	if err == sql.ErrNoRows {
		internal.RespondWithError(w, http.StatusNotFound, internal.ErrGameNotFound)
		return
	} else if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}

	internal.RespondWithData(w, g)
}

// UpdateGame updates the day of a game.
func UpdateGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	var payload struct {
		Day int `json:"day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrValidationFailed)
		return
	}

	const q = `UPDATE games SET day = $1 WHERE id = $2`
	res, err := internal.DB.Exec(q, payload.Day, id)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		internal.RespondWithError(w, http.StatusNotFound, internal.ErrGameNotFound)
		return
	}

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// DeleteGame deletes a game by ID.
func DeleteGame(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	const q = `DELETE FROM games WHERE id = $1`
	res, err := internal.DB.Exec(q, id)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		internal.RespondWithError(w, http.StatusNotFound, internal.ErrGameNotFound)
		return
	}

	// 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// GetEvents returns the audit log for a game.
func GetEvents(w http.ResponseWriter, r *http.Request) {
	gameID, err := uuid.Parse(r.URL.Query().Get("game_id"))
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}
	rows, err := internal.DB.Query(
		`SELECT id, card_id, event_type, payload, created_at 
           FROM game_events 
          WHERE game_id=$1 
       ORDER BY created_at`,
		gameID,
	)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
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
	internal.RespondWithData(w, events)
}

// GetBoard returns all cards for a game, grouped by column.
func GetBoard(w http.ResponseWriter, r *http.Request) {
	gameIDStr := r.URL.Query().Get("game_id")
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	rows, err := internal.DB.Query(`
        SELECT
            id, game_id, title, card_column,
            class_of_service, value_estimate,
            effort_analysis, effort_development, effort_test,
            selected_day, deployed_day
        FROM cards
        WHERE game_id = $1
        ORDER BY created_at
    `, gameID)
	if err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}
	defer rows.Close()

	board := map[string][]Card{}
	for rows.Next() {
		var c Card
		if err := rows.Scan(
			&c.ID,
			&c.GameID,
			&c.Title,
			&c.CardColumn,
			&c.ClassOfService,
			&c.ValueEstimate,
			&c.EffortAnalysis,
			&c.EffortDev,
			&c.EffortTest,
			&c.SelectedDay,
			&c.DeployedDay,
		); err != nil {
			internal.RespondWithError(w, http.StatusInternalServerError, internal.ErrDatabaseError)
			return
		}
		board[c.CardColumn] = append(board[c.CardColumn], c)
	}

	internal.RespondWithData(w, board)
}
