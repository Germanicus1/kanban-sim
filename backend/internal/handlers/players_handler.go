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
// @Description.markdown player_create
// @Tags         players
// @Accept       json
// @Produce      json
// @Param        payload  body      models.CreatePlayerRequest  true  "Player creation payload"
// @Success      200      {string}  string                     "Created player UUID"
// @Failure      400      {object}  response.ErrorResponse     "Invalid game ID or player name"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      405      {object}  response.ErrorResponse     "Method not allowed"
// @Failure      500      {object}  response.ErrorResponse     "Internal server error"
// @Security    BearerAuth
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

// GetPlayerByID retrieves a player by UUID.
// @Summary      Get player by ID
// @Description  Returns the full player record for the given player UUID.
// @Tags         players
// @Produce      json
// @Param        id   path      string  true  "Game ID" Format(uuid)
// @Success      200  {object}  models.Player            "Player retrieved successfully"
// @Failure      400  {object}  response.ErrorResponse   "Invalid or missing player ID"
// @Failure      403  {object}  response.ErrorResponse   "Missing or invalid token"
// @Failure      404  {object}  response.ErrorResponse   "Player not found"
// @Failure      405  {object}  response.ErrorResponse   "Method not allowed"
// @Failure      500  {object}  response.ErrorResponse   "Internal server error"
// @Security    BearerAuth
// @Router       /players/{id} [get]
func (h *PlayerHandler) GetPlayerByID(w http.ResponseWriter, r *http.Request) {
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
	playerID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
		return
	}

	player, err := h.Service.GetPlayerByID(r.Context(), playerID)
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

// @Summary      Update a player
// @Description.markdown player_update
// @Tags         players
// @Accept       json
// @Produce      json
// @Param        payload  body      models.UpdatePlayerRequest  true  "Player update payload"
// @Success      200      {string}  string                      "Update successful (empty response)"
// @Failure      400      {object}  response.ErrorResponse     "Invalid player ID or name"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      405      {object}  response.ErrorResponse     "Method not allowed"
// @Failure      500      {object}  response.ErrorResponse     "Internal server error"
// @Security    BearerAuth
// @Router       /players [patch]
func (h *PlayerHandler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	var payload models.UpdatePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrValidationFailed)
		return
	}
	if payload.ID == uuid.Nil || payload.Name == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerData)
		return
	}

	if err := h.Service.UpdatePlayer(r.Context(), payload.ID, payload.Name); err != nil {
		status, code := response.MapPostgresError(err)
		response.RespondWithError(w, status, code)
		return
	}

	response.RespondWithData(w, "")
}

// DeletePlayer removes a player by UUID.
// @Summary      Delete a player
// @Description  Deletes the player identified by the given UUID.
// @Tags         players
// @Accept       json
// @Produce      json
// @Param        payload  body      models.DeletePlayerRequest  true  "Player deletion payload"
// @Success      200      {string}  string                      "Player deleted successfully"
// @Failure      400      {object}  response.ErrorResponse      "Invalid or missing player ID"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      404      {object}  response.ErrorResponse      "Player not found"
// @Failure      405      {object}  response.ErrorResponse      "Method not allowed"
// @Failure      500      {object}  response.ErrorResponse      "Internal server error"
// @Security    BearerAuth
// @Router       /players [delete]
func (h *PlayerHandler) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	var payload models.DeletePlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrValidationFailed)
		return
	}
	if payload.ID == uuid.Nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidPlayerID)
		return
	}
	if err := h.Service.DeletePlayer(r.Context(), payload.ID); err != nil {
		if errors.Is(err, players.ErrNotFound) {
			response.RespondWithError(w, http.StatusNotFound, response.ErrPlayerNotFound)
		} else {
			status, code := response.MapPostgresError(err)
			response.RespondWithError(w, status, code)
		}
		return
	}
	response.RespondWithData(w, "Player deleted successfully")
}

// ListPlayersByGameID retrieves all players for a specific game.
// @Summary      List all players by game ID
// @Description  Returns a list of players belonging to the given game UUID.
// @Tags         players
// @Produce      json
// @Param        game_id   path      string           true  "Game ID"  Format(uuid)
// @Success      200       {array}   models.Player    "List of players"
// @Failure      400       {object}  response.ErrorResponse  "Invalid or missing game ID"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      404       {object}  response.ErrorResponse  "Players not found"
// @Failure      405       {object}  response.ErrorResponse  "Method not allowed"
// @Failure      500       {object}  response.ErrorResponse  "Internal server error"
// @Security    BearerAuth
// @Router       /games/{game_id}/players [get]
func (h *PlayerHandler) ListPlayersByGameID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}
	gameIDStr := r.PathValue("game_id")
	if gameIDStr == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	players, err := h.Service.ListPlayersByGameID(r.Context(), gameID)
	if err != nil {
		if errors.Is(err, errors.New(response.ErrPlayersNotFound)) {
			response.RespondWithError(w, http.StatusNotFound, response.ErrPlayersNotFound)
		} else {
			status, code := response.MapPostgresError(err)
			response.RespondWithError(w, status, code)
		}
		return
	}
	response.RespondWithData(w, players)
}
