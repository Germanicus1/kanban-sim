package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/response"
	"github.com/google/uuid"
)

type Player struct {
	ID       uuid.UUID `json:"id"`
	GameID   uuid.UUID `json:"game_id"`
	Name     string    `json:"name"`
	JoinedAt string    `json:"joined_at,omitempty"`
}

// CreatePlayer creates a new player
func CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var p Player
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrValidationFailed)
		return
	}
	if p.GameID == uuid.Nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
		return
	}

	const q = `
        INSERT INTO players (game_id, name)
        VALUES ($1, $2)
        RETURNING id
    `
	if err := database.DB.QueryRow(q, p.GameID, p.Name).Scan(&p.ID); err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}

	response.RespondWithData(w, p)
}

// GetPlayer retrieves a player by ID
func GetPlayer(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
		return
	}

	const q = `
        SELECT id, game_id, name
          FROM players
         WHERE id = $1
    `
	var p Player
	err = database.DB.QueryRow(q, id).Scan(&p.ID, &p.GameID, &p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			response.RespondWithError(w, http.StatusNotFound, response.ErrPlayerNotFound)
		} else {
			status, code := response.MapPostgresError(err)
			response.RespondWithError(w, status, code)
		}
		return
	}

	response.RespondWithData(w, p)
}

// UpdatePlayer updates a player's name
func UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
		return
	}

	var payload struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrValidationFailed)
		return
	}

	const q = `UPDATE players SET name = $1 WHERE id = $2`
	res, err := database.DB.Exec(q, payload.Name, id)
	if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		response.RespondWithError(w, http.StatusNotFound, response.ErrPlayerNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePlayer deletes a player by ID
func DeletePlayer(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
		return
	}

	const q = `DELETE FROM players WHERE id = $1`
	res, err := database.DB.Exec(q, id)
	if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		response.RespondWithError(w, http.StatusNotFound, response.ErrPlayerNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
