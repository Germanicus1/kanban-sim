package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal/config"
	"github.com/Germanicus1/kanban-sim/internal/games"
	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/Germanicus1/kanban-sim/internal/response"
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

// CreateGame handles POST /games by loading the embedded config
// and passing it straight into your Service.
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
				ClassOfService: safeString(cc.ClassOfService),
				ValueEstimate:  safeString(cc.ValueEstimate),
				SelectedDay:    safeInt(cc.SelectedDay),
				DeployedDay:    safeInt(cc.DeployedDay),
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

// helper to turn *string → string
func safeString(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return "" // or some default value
}

// safeInt turns a *int into an int, using 0 if the pointer is nil.
func safeInt(ptr *int) int {
	if ptr != nil {
		return *ptr
	}
	return 0
}

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
		response.RespondWithError(w, http.StatusInternalServerError, err.Error())
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
		if errors.Is(err, games.ErrNotFound) {
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
		if errors.Is(err, games.ErrNotFound) {
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
