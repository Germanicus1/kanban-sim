package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
)

type Player struct {
	ID       uuid.UUID `json:"id"`
	GameID   uuid.UUID `json:"game_id"`
	Name     string    `json:"name"`
	JoinedAt string    `json:"joined_at"`
}

// CreatePlayer creates a new player
func CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var player Player

	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if player.GameID == uuid.Nil {
		http.Error(w, "GameID is required", http.StatusBadRequest)
		return
	}

	if player.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO players (game_id, name) 
		VALUES ($1, $2) 
		RETURNING id, joined_at
	`
	err := internal.DB.QueryRow(query, player.GameID, player.Name).Scan(&player.ID, &player.JoinedAt)
	if err != nil {
		http.Error(w, "Failed to create player", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(player)
}

// GetPlayer retrieves a player by ID
func GetPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	var player Player
	query := `SELECT id, game_id, name, joined_at FROM players WHERE id = $1`
	err := internal.DB.QueryRow(query, id).Scan(&player.ID, &player.GameID, &player.Name, &player.JoinedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch player", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(player)
}

// UpdatePlayer updates a player's name
func UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	var data struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `UPDATE players SET name = $1 WHERE id = $2`
	_, err := internal.DB.Exec(query, data.Name, id)
	if err != nil {
		http.Error(w, "Failed to update player", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePlayer deletes a player by ID
func DeletePlayer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM players WHERE id = $1`
	result, err := internal.DB.Exec(query, id)
	if err != nil {
		http.Error(w, "Failed to delete player", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
