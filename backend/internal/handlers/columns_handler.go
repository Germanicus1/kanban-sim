package handlers

import (
	"errors"
	"net/http"

	"github.com/Germanicus1/kanban-sim/backend/internal/columns"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

type ColumnsHandler struct {
	Service columns.ColumnServiceInterface
}

func NewColumnHandler(svc columns.ColumnServiceInterface) *ColumnsHandler {
	return &ColumnsHandler{Service: svc}
}

// GetColumnsByGameID retrieves all columns for a specific game.
// @Summary      List columns by game ID
// @Description  Returns the list of columns (including subcolumns) belonging to the specified game UUID.
// @Tags         columns
// @Produce      json
// @Param        id   path      string           true  "Game ID"   Format(uuid)
// @Success      200  {array}   models.Column    "List of columns"
// @Failure      400  {object}  response.ErrorResponse  "Invalid or missing game ID"
// @Failure      404  {object}  response.ErrorResponse  "Game not found"
// @Failure      405  {object}  response.ErrorResponse  "Method not allowed"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Router       /games/{id}/columns [get]
func (h *ColumnsHandler) GetColumnsByGameID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}
	idStr := r.PathValue("id")
	if idStr == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	columns, err := h.Service.GetColumnsByGameID(r.Context(), gameID)
	if err != nil {
		if errors.Is(err, errors.New(response.ErrPlayerNotFound)) {
			response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		} else {
			status, code := response.MapPostgresError(err)
			response.RespondWithError(w, status, code)
		}
		return
	}
	response.RespondWithData(w, columns)
}
