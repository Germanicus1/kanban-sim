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

// CreateGame creates a new game
func CreateGame(w http.ResponseWriter, r *http.Request) {
	var game Game
	game.Day = 1

	query := `
		INSERT INTO games (day, columns) 
		VALUES ($1, '[]'::jsonb) 
		RETURNING id, created_at
	`
	err := internal.DB.QueryRow(query, game.Day).Scan(&game.ID, &game.CreatedAt)
	if err != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, internal.ErrDatabaseError)
		return
	}

	internal.RespondWithData(w, game)
}

// GetGame retrieves a game by ID
func GetGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	var game Game
	query := `SELECT id, created_at, day, columns FROM games WHERE id = $1`
	err = internal.DB.QueryRow(query, parsedID).Scan(&game.ID, &game.CreatedAt, &game.Day, &game.Columns)
	if err == sql.ErrNoRows {
		internal.RespondWithError(w, http.StatusNotFound, internal.ErrGameNotFound)
		return
	} else if err != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, internal.ErrDatabaseError)
		return
	}

	internal.RespondWithData(w, game)
}

// UpdateGame updates the day of a game
func UpdateGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	var data struct {
		Day int `json:"day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrValidationFailed)
		return
	}

	query := `UPDATE games SET day = $1 WHERE id = $2`
	_, err = internal.DB.Exec(query, data.Day, parsedID)
	if err != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, internal.ErrDatabaseError)
		return
	}

	internal.RespondWithData(w, map[string]string{
		"message": "Game updated successfully",
	})
}

// DeleteGame deletes a game by ID
func DeleteGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	parsedID, err := uuid.Parse(id)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, internal.ErrInvalidGameID)
		return
	}

	query := `DELETE FROM games WHERE id = $1`
	result, err := internal.DB.Exec(query, parsedID)
	if err != nil {
		internal.RespondWithError(w, http.StatusInternalServerError, internal.ErrDatabaseError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		internal.RespondWithError(w, http.StatusNotFound, internal.ErrGameNotFound)
		return
	}

	internal.RespondWithData(w, map[string]string{
		"message": "Game deleted successfully",
	})
}
