package response

import (
	"encoding/json"
	"net/http"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
)

// APIResponse is the standard envelope for all API responses.
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CreateGameData is the payload for a newly created game.
type CreateGameData struct {
	ID string `json:"id" example:"7d7881cf-8d9f-457f-ac93-aa498ea8c0af"`
}

// CreateGameResponse is the full envelope returned by CreateGame.
// swagger:model CreateGameResponse
type CreateGameResponse struct {
	Success bool           `json:"success"`
	Data    CreateGameData `json:"data"`
}

// ErrorResponse is the shape used for errors.
// swagger:model ErrorResponse
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"contextual error message"`
}

// GameResponse is the envelope for a single game.
// swagger:model GameResponse
type GameResponse struct {
	Success bool        `json:"success" example:"true"`
	Data    models.Game `json:"data"`
}

// RespondWithError writes a JSON error response.
func RespondWithError(w http.ResponseWriter, status int, errCode string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse[any]{
		Success: false,
		Error:   errCode,
	})
}

// RespondWithData writes a JSON success response.
func RespondWithData(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse[any]{
		Success: true,
		Data:    data,
	})
}
