package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Germanicus1/kanban-sim/backend/internal/config"
	"github.com/Germanicus1/kanban-sim/backend/internal/games"
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

// GameHandler groups your game endpoints.
type GameHandler struct {
	Service games.ServiceInterface
}

type updateGameRequest struct {
	Day int `json:"day"`
}

// NewGameHandler constructs a GameHandler.
func NewGameHandler(svc games.ServiceInterface) *GameHandler {
	return &GameHandler{Service: svc}
}

// CreateGame creates a new game with the default board.
// @Summary      Create a new game
// @Description  Creates a new game using the embedded default board; no request body required
// @Tags         games
// @Produce      json
// @Success      201  {object}  response.CreateGameResponse "New game created"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Router       /games [post]
func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	// Only accept POST
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load the board config from your embedded JSON
	cfg, err := config.LoadBoardConfig()
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"failed to load board config: "+err.Error())
		return
	}

	gameCfg := models.BoardConfig{
		EffortTypes: cfg.EffortTypes,
		Columns:     cfg.Columns,
		Cards:       make([]models.Card, len(cfg.Cards)),
	}

	for i, cc := range cfg.Cards {
		_ = i
		// map efforts
		efforts := make([]models.Effort, len(cc.Efforts))
		for j, ce := range cc.Efforts {
			efforts[j] = models.Effort{
				EffortType: ce.EffortType,
				Estimate:   ce.Estimate,
			}
		}

		gameCfg.Cards = make([]models.Card, len(cfg.Cards))
		for i, cc := range cfg.Cards {

			gameCfg.Cards[i] = models.Card{
				ColumnTitle:    cc.ColumnTitle,
				Title:          cc.Title,
				ClassOfService: cc.ClassOfService,
				ValueEstimate:  cc.ValueEstimate,
				SelectedDay:    cc.SelectedDay,
				DeployedDay:    cc.DeployedDay,
				Efforts:        efforts,
			}
		}
	}

	// Now call your service with the correctly‐typed gameCfg
	gameID, err := h.Service.CreateGame(r.Context(), gameCfg)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"could not create game: "+err.Error())
		return
	}

	// 5) Return the new game ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response.RespondWithData(w, map[string]string{"id": gameID.String()})
}

// GetGame retrieves a game by its UUID.
// @Summary      Get game by ID
// @Description  Returns the full game record for the given UUID.
// @Tags         games
// @Produce      json
// @Param        id   path      string  true  "Game ID" Format(uuid)
// @Success      200  {object}  response.GameResponse   "Game retrieved successfully"
// @Failure      400  {object}  response.ErrorResponse  "Invalid or missing game ID"
// @Failure      404  {object}  response.ErrorResponse  "Game not found"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Router       /games/{id} [get]
func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
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

	game, err := h.Service.GetGame(r.Context(), gameID)
	if err != nil {
		log.Printf("GetGame: failed to load game for %s: %v", gameID, err)
		response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		return
	}

	response.RespondWithData(w, game)
}

// GetBoard handles GET /games/{id}/board
func (h *GameHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	// 1) Method check (optional, since mux already matched “GET”)
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	// 2) Extract the {id} wildcard from the path
	idStr := r.PathValue("id")
	if idStr == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	// 3) Parse it as a UUID
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	// 4) Delegate to your service
	board, err := h.Service.GetBoard(r.Context(), gameID)
	if err != nil {
		log.Printf("GetBoard: failed to load board for %s: %v", gameID, err)
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
	}

	response.RespondWithData(w, board)
	// 5) Encode and return
	// w.Header().Set("Content-Type", "application/json")
	// w.WriteHeader(http.StatusOK)
	// json.NewEncoder(w).Encode(board)
}

// internal/handlers/game_handler.go
func (h *GameHandler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	// TODO: Route for DELETE /games/{id}
	// This works because of the route "DELETE /games/{id}"
	idStr := r.PathValue("id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	if err := h.Service.DeleteGame(r.Context(), gameID); err != nil {
		if errors.Is(err, response.ErrNotFound) {
			response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		} else {
			response.RespondWithError(w, http.StatusInternalServerError, response.ErrInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *GameHandler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	var req updateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidJSON)
		return
	}

	if err := h.Service.UpdateGame(r.Context(), gameID, req.Day); err != nil {
		if errors.Is(err, response.ErrNotFound) {
			response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		} else {
			response.RespondWithError(w, http.StatusInternalServerError, response.ErrInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/*
curl -i -X PATCH \
  http://localhost:8080/games/4e0a45df-b1e7-4c9c-9ff2-c063fc65e2e5 \
  -H "Content-Type: application/json" \
  -d '{"day":5}'


*/
