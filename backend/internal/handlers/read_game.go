package handlers

import (
	"database/sql"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/Germanicus1/kanban-sim/internal/response"
	"github.com/google/uuid"
)

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
