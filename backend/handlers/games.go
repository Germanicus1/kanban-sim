package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
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

	query := `INSERT INTO games (day, columns) VALUES ($1, '[]'::jsonb) RETURNING id, created_at`
	err := internal.DB.QueryRow(query, game.Day).Scan(&game.ID, &game.CreatedAt)
	if err != nil {
		http.Error(w, "Failed to create game", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)
}

// GetGame retrieves a game by ID
func GetGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var game Game
	query := `SELECT id, created_at, day, columns FROM games WHERE id = $1`
	err := internal.DB.QueryRow(query, id).Scan(&game.ID, &game.CreatedAt, &game.Day, &game.Columns)
	if err == sql.ErrNoRows {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch game", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(game)
}

// UpdateGame updates the day of a game
func UpdateGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	var data struct {
		Day int `json:"day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `UPDATE games SET day = $1 WHERE id = $2`
	_, err := internal.DB.Exec(query, data.Day, id)
	if err != nil {
		http.Error(w, "Failed to update game", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteGame deletes a game by ID
// DeleteGame deletes a game by ID
func DeleteGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to delete game with ID: %s", id)

	query := `DELETE FROM games WHERE id = $1`
	result, err := internal.DB.Exec(query, id)
	if err != nil {
		log.Printf("SQL error: %v", err)
		http.Error(w, "Failed to delete game", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Game not found for ID: %s", id)
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	log.Printf("Successfully deleted game with ID: %s", id)
	w.WriteHeader(http.StatusNoContent)
}
