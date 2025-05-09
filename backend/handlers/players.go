package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
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
// CreatePlayer creates a new player
func CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var player Player

	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		log.Printf("Error decoding request body: %v", err)
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY")
		return
	}

	if player.GameID == uuid.Nil {
		log.Println("Game ID is required")
		internal.RespondWithError(w, http.StatusBadRequest, "GAME_ID_REQUIRED")
		return
	}

	query := `
		INSERT INTO players (game_id, name) 
		VALUES ($1, $2) 
		RETURNING id
	`

	err := internal.DB.QueryRow(query, player.GameID, player.Name).Scan(&player.ID)
	if err != nil {
		status, errCode := internal.MapPostgresError(err)
		log.Printf("Error inserting player: %v", err)
		internal.RespondWithError(w, status, errCode)
		return
	}

	internal.RespondWithData(w, player)
}

// GetPlayer retrieves a player by ID
func GetPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid Player ID: %s", id)
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_PLAYER_ID")
		return
	}

	log.Printf("Looking for player with ID: %s", parsedID)

	var player Player
	query := `SELECT id, game_id, name FROM players WHERE id = $1`

	err = internal.DB.QueryRow(query, parsedID).Scan(&player.ID, &player.GameID, &player.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Player not found with ID: %s", parsedID)
			internal.RespondWithError(w, http.StatusNotFound, "PLAYER_NOT_FOUND")
		} else {
			log.Printf("Database error: %v", err)
			internal.RespondWithError(w, http.StatusInternalServerError, "DATABASE_ERROR")
		}
		return
	}

	internal.RespondWithData(w, player)
}

// UpdatePlayer updates a player's name
func UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	parsedID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid Player ID: %s", id)
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_PLAYER_ID")
		return
	}

	log.Printf("Updating player with ID: %s", parsedID)

	var data struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Error decoding update payload: %v", err)
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY")
		return
	}

	query := `UPDATE players SET name = $1 WHERE id = $2`
	result, err := internal.DB.Exec(query, data.Name, parsedID)
	if err != nil {
		log.Printf("Database error during update: %v", err)
		internal.RespondWithError(w, http.StatusInternalServerError, "DATABASE_ERROR")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("Player not found for update: %s", parsedID)
		internal.RespondWithError(w, http.StatusNotFound, "PLAYER_NOT_FOUND")
		return
	}

	log.Printf("Player updated: %s to %s", parsedID, data.Name)

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
