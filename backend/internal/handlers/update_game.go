package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Germanicus1/kanban-sim/backend/internal/database"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

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
