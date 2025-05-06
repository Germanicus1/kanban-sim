package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/Germanicus1/kanban-sim/internal"
)

func GameRouter(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/game/")
	parts := strings.Split(path, "/")

	switch {
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "next-day":
		handleNextDay(w, r, parts[0])
	case r.Method == "GET" && len(parts) == 1:
		handleGetGame(w, r)
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "join":
		handleJoinGame(w, r, parts[0])
	case r.Method == "GET" && len(parts) == 2 && parts[1] == "players":
		handleGetPlayers(w, r, parts[0])
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "leave":
		handleLeaveGame(w, r, parts[0])
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "end":
		handleEndGame(w, r, parts[0])
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "reset":
		handleResetGame(w, r, parts[0])
	default:
		http.NotFound(w, r)
	}
}

func handleGetGame(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/game/")
	if id == "" {
		http.Error(w, "Missing game ID", http.StatusBadRequest)
		return
	}

	resp, _, err := internal.Supabase.
		From("games").
		Select("*", "exact", false).
		Eq("id", id).
		Execute()

	if err != nil {
		http.Error(w, "Failed to fetch game", http.StatusInternalServerError)
		log.Println("Supabase fetch error:", err)
		return
	}

	w.Write(resp)
}

func handleNextDay(w http.ResponseWriter, r *http.Request, gameID string) {
	resp, _, err := internal.Supabase.
		From("games").
		Select("*", "exact", false).
		Eq("id", gameID).
		Execute()
	if err != nil {
		http.Error(w, "Failed to fetch game", 500)
		return
	}

	var games []map[string]interface{}
	if err := json.Unmarshal(resp, &games); err != nil || len(games) == 0 {
		http.Error(w, "Game not found", 404)
		return
	}
	day := int(games[0]["day"].(float64)) + 1

	updated, _, err := internal.Supabase.
		From("games").
		Update(map[string]interface{}{"day": day}, "", "").
		Eq("id", gameID).
		Execute()
	if err != nil {
		http.Error(w, "Failed to update day", 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(updated)
}
