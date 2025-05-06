package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Germanicus1/kanban-sim/internal"
)

func CardRouter(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/cards/")
	parts := strings.Split(path, "/")

	if r.Method == "GET" && len(parts) == 1 {
		handleGetCards(w, r, parts[0])
		return
	}

	if r.Method == "POST" && len(parts) == 2 && parts[1] == "move" {
		handleMoveCard(w, r, parts[0])
		return
	}

	http.NotFound(w, r)
}

func handleGetCards(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	resp, _, err := internal.Supabase.
		From("cards").
		Select("*", "exact", false).
		Eq("game_id", gameID).
		Execute()

	if err != nil {
		http.Error(w, "Failed to fetch cards", 500)
		log.Println("Fetch cards error:", err)
		return
	}

	w.Write(resp)
}

func handleMoveCard(w http.ResponseWriter, r *http.Request, cardID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	var input struct {
		NewColumn string `json:"new_column"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid body", 400)
		return
	}

	_, _, err := internal.Supabase.
		From("cards").
		Update(map[string]interface{}{"card_column": input.NewColumn}, "", "").
		Eq("id", cardID).
		Execute()

	if err != nil {
		http.Error(w, "Failed to move card", 500)
		log.Println("Move card error:", err)
		return
	}

	w.Write([]byte(`{"status":"ok"}`))
}
