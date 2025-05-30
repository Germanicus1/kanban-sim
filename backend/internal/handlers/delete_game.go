package handlers

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/backend/internal/database"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

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
