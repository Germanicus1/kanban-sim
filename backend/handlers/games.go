package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
)

type Game struct {
	ID        uuid.UUID       `json:"id"`
	CreatedAt string          `json:"created_at"`
	Day       int             `json:"day"`
	Columns   json.RawMessage `json:"columns"`
}

// CreateGame creates a new game.
func CreateGame(w http.ResponseWriter, r *http.Request) {
	var g Game
	g.Day = 1

	const q = `
		INSERT INTO games (day, columns)
		VALUES ($1, '[]'::jsonb)
		RETURNING id, created_at
	`
	if err := internal.DB.QueryRow(q, g.Day).Scan(&g.ID, &g.CreatedAt); err != nil {
		status, code := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, code)
		return
	}

	internal.RespondWithData(w, g)
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
