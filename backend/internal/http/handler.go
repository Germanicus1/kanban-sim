package http

import (
	"encoding/json"
	"net/http"
)

// Handler holds references to services
type Handler struct {
	GameSvc games.GameService
}

// NewHandler wires services into HTTP handlers
func NewHandler(svc games.GameService) *Handler {
	return &Handler{GameSvc: svc}
}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {
	var input games.CreateGameInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request payload")
		return
	}
	ctx := r.Context()
	game, err := h.GameSvc.CreateGame(ctx, input)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, game)
}

// ...other handlers wired the same way
