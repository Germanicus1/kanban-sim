package internal

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func RespondWithError(w http.ResponseWriter, status int, errCode string) {
	response := APIResponse{
		Success: false,
		Error:   errCode,
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func RespondWithData(w http.ResponseWriter, data interface{}) {
	response := APIResponse{
		Success: true,
		Data:    data,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
