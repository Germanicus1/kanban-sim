package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal"

	"github.com/supabase-community/postgrest-go"
)

// HandleJoinGame adds a player to the game
func HandleJoinGame(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		http.Error(w, "Missing player name", 400)
		return
	}

	player := []map[string]interface{}{
		{
			"game_id": gameID,
			"name":    input.Name,
		},
	}

	resp, _, err := internal.Supabase.
		From("players").
		Insert(player, false, "", "representation", "").
		Execute()

	if err != nil {
		http.Error(w, "Failed to join game", 500)
		log.Println("Join game error:", err)
		return
	}

	w.Write(resp)
}

// HandleGetPlayers returns all players in a game
func HandleGetPlayers(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	resp, _, err := internal.Supabase.
		From("players").
		Select("*", "exact", false).
		Eq("game_id", gameID).
		Order("joined_at", &postgrest.OrderOpts{Ascending: true}).
		Execute()

	if err != nil {
		http.Error(w, "Failed to fetch players", 500)
		log.Println("Get players error:", err)
		return
	}

	w.Write(resp)
}

// HandleLeaveGame removes a player from the game
func HandleLeaveGame(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	var input struct {
		PlayerID string `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.PlayerID == "" {
		http.Error(w, "Missing player ID", 400)
		return
	}

	_, _, err := internal.Supabase.
		From("players").
		Delete("", "").
		Eq("game_id", gameID).
		Eq("id", input.PlayerID).
		Execute()
	if err != nil {
		http.Error(w, "Failed to leave game", 500)
		log.Println("Leave game error:", err)
		return
	}

	w.Write([]byte(`{"status":"left"}`))
}
