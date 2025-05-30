package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/Germanicus1/kanban-sim/backend/internal/players"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

// GameHandler groups your player endpoints.
type PlayerHandler struct {
	Service players.ServiceInterface
}

func NewPlayerHandler(svc players.ServiceInterface) *PlayerHandler {
	return &PlayerHandler{Service: svc}
}

type Player struct {
	ID       uuid.UUID `json:"id"`
	GameID   uuid.UUID `json:"game_id"`
	Name     string    `json:"name"`
	JoinedAt string    `json:"joined_at,omitempty"`
}

// CreatePlayer creates a new player in a game.
// @Summary      Create a new player
// @Description  Creates a player under the specified game ID and returns the new player's UUID.
// @Tags         players
// @Accept       json
// @Produce      json
// @Param        request  body      players.CreatePlayerRequest  true  "Player creation payload"
// @Success      200      {string}  string                      "Created player UUID"
// @Failure      400      {object}  response.ErrorResponse      "Validation failed or invalid JSON"
// @Failure      405      {object}  response.ErrorResponse      "Method not allowed"
// @Failure      500      {object}  response.ErrorResponse      "Internal server error"
// @Router       /players [post]
func (h *PlayerHandler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	// Only accept POST
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload models.CreatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrValidationFailed)
		return
	}
	if payload.GameID == uuid.Nil || payload.Name == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerData)
		return
	}

	player, err := h.Service.CreatePlayer(r.Context(), payload.GameID, payload.Name)
	if err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}

	response.RespondWithData(w, player)
}

func (h *PlayerHandler) GetPlayerByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
		return
	}

	player, err := h.Service.GetPlayerByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, errors.New(response.ErrPlayerNotFound)) {
			response.RespondWithError(w, http.StatusNotFound, response.ErrPlayerNotFound)
		} else {
			status, code := response.MapPostgresError(err)
			response.RespondWithError(w, status, code)
		}
		return
	}

	response.RespondWithData(w, player)
}

// // GetPlayer retrieves a player by ID
// func GetPlayer(w http.ResponseWriter, r *http.Request) {
// 	idStr := r.URL.Query().Get("id")
// 	id, err := uuid.Parse(idStr)
// 	if err != nil {
// 		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
// 		return
// 	}

// 	const q = `
//         SELECT id, game_id, name
//           FROM players
//          WHERE id = $1
//     `
// 	var p Player
// 	err = database.DB.QueryRow(q, id).Scan(&p.ID, &p.GameID, &p.Name)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			response.RespondWithError(w, http.StatusNotFound, response.ErrPlayerNotFound)
// 		} else {
// 			status, code := response.MapPostgresError(err)
// 			response.RespondWithError(w, status, code)
// 		}
// 		return
// 	}

// 	response.RespondWithData(w, p)
// }

// // UpdatePlayer updates a player's name
// func UpdatePlayer(w http.ResponseWriter, r *http.Request) {
// 	idStr := r.URL.Query().Get("id")
// 	id, err := uuid.Parse(idStr)
// 	if err != nil {
// 		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
// 		return
// 	}

// 	var payload struct {
// 		Name string `json:"name"`
// 	}
// 	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
// 		response.RespondWithError(w, http.StatusBadRequest, response.ErrValidationFailed)
// 		return
// 	}

// 	const q = `UPDATE players SET name = $1 WHERE id = $2`
// 	res, err := database.DB.Exec(q, payload.Name, id)
// 	if err != nil {
// 		status, code := response.MapPostgresError(err)
// 		response.RespondWithError(w, status, code)
// 		return
// 	}
// 	if n, _ := res.RowsAffected(); n == 0 {
// 		response.RespondWithError(w, http.StatusNotFound, response.ErrPlayerNotFound)
// 		return
// 	}

// 	w.WriteHeader(http.StatusNoContent)
// }

// // DeletePlayer deletes a player by ID
// func DeletePlayer(w http.ResponseWriter, r *http.Request) {
// 	idStr := r.URL.Query().Get("id")
// 	id, err := uuid.Parse(idStr)
// 	if err != nil {
// 		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
// 		return
// 	}

// 	const q = `DELETE FROM players WHERE id = $1`
// 	res, err := database.DB.Exec(q, id)
// 	if err != nil {
// 		status, code := response.MapPostgresError(err)
// 		response.RespondWithError(w, status, code)
// 		return
// 	}
// 	if n, _ := res.RowsAffected(); n == 0 {
// 		response.RespondWithError(w, http.StatusNotFound, response.ErrPlayerNotFound)
// 		return
// 	}

// 	w.WriteHeader(http.StatusNoContent)
// }
