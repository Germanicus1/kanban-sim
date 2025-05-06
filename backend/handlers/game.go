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
		HandleJoinGame(w, r, parts[0])
	case r.Method == "GET" && len(parts) == 2 && parts[1] == "players":
		HandleGetPlayers(w, r, parts[0])
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "leave":
		HandleLeaveGame(w, r, parts[0])
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "end":
		HandleEndGame(w, r, parts[0])
	case r.Method == "POST" && len(parts) == 2 && parts[1] == "reset":
		HandleResetGame(w, r, parts[0])
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

func HandleEndGame(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	// Delete players
	_, _, err := internal.Supabase.From("players").Delete("", "").Eq("game_id", gameID).Execute()
	if err != nil {
		http.Error(w, "Failed to delete players", 500)
		log.Println("Delete players error:", err)
		return
	}

	// Delete cards
	_, _, err = internal.Supabase.From("cards").Delete("", "").Eq("game_id", gameID).Execute()
	if err != nil {
		http.Error(w, "Failed to delete cards", 500)
		log.Println("Delete cards error:", err)
		return
	}

	// Insert game event: ended
	_, _, err = internal.Supabase.From("game_events").Insert([]map[string]interface{}{
		{
			"game_id": gameID,
			"type":    "ended",
		},
	}, false, "", "representation", "").Execute()
	if err != nil {
		log.Println("Insert game event failed:", err)
	}

	// Delete game
	_, _, err = internal.Supabase.From("games").Delete("", "").Eq("id", gameID).Execute()
	if err != nil {
		http.Error(w, "Failed to delete game", 500)
		log.Println("Delete game error:", err)
		return
	}

	w.Write([]byte(`{"status":"ended"}`))
}

func HandleResetGame(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	config, err := internal.LoadBoardConfig("config/board_config.json")
	if err != nil {
		http.Error(w, "Failed to load board config", 500)
		log.Println("LoadBoardConfig error:", err)
		return
	}

	// Delete players and cards
	for _, table := range []string{"players", "cards"} {
		_, _, err := internal.Supabase.From(table).Delete("", "").Eq("game_id", gameID).Execute()
		if err != nil {
			http.Error(w, "Failed to delete "+table, 500)
			log.Printf("Delete %s error: %v\n", table, err)
			return
		}
	}

	// Update game columns and day
	_, _, err = internal.Supabase.
		From("games").
		Update(map[string]interface{}{
			"columns": config.Columns,
			"day":     1,
		}, "", "").
		Eq("id", gameID).
		Execute()
	if err != nil {
		http.Error(w, "Failed to update game columns", 500)
		log.Println("Update game columns error:", err)
		return
	}

	// Insert cards
	var cardsToInsert []map[string]interface{}
	for _, card := range config.InitialCards {
		entry := map[string]interface{}{
			"game_id":            gameID,
			"title":              card.ID,
			"card_column":        card.ColumnID,
			"class_of_service":   card.ClassOfService,
			"value_estimate":     card.ValueEstimate,
			"effort_analysis":    card.Effort.Analysis,
			"effort_development": card.Effort.Development,
			"effort_test":        card.Effort.Test,
			"selected_day":       card.SelectedDay,
			"deployed_day":       card.DeployedDay,
		}
		cardsToInsert = append(cardsToInsert, entry)
	}

	_, _, err = internal.Supabase.
		From("cards").
		Insert(cardsToInsert, false, "", "representation", "").
		Execute()
	if err != nil {
		http.Error(w, "Failed to insert reset cards", 500)
		log.Println("Insert reset cards error:", err)
		return
	}

	w.Write([]byte(`{"status":"reset complete"}`))
}
