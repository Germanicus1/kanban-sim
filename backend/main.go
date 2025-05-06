package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/joho/godotenv"
	"github.com/supabase-community/postgrest-go"
	supa "github.com/supabase-community/supabase-go"
)

var supabase *supa.Client

func main() {
	// Load .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env not found or couldn't load")
	}

	// Init Supabase
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	supabase, err = supa.NewClient(supabaseURL, supabaseKey, nil)
	if err != nil {
		log.Fatalf("Failed to init Supabase: %v", err)
	}

	// Routes
	http.HandleFunc("/ping", handlePing)
	http.HandleFunc("/create-game", handleCreateGame)
	http.HandleFunc("/cards/", cardRouter)
	http.HandleFunc("/game/", gameRouter)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	fmt.Fprintln(w, "Backend running")
}

func handleCreateGame(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	game := []map[string]interface{}{
		{
			"day":     1,
			"columns": []string{"Backlog", "Dev", "Test", "Done"},
		},
	}

	resp, _, err := supabase.From("games").Insert(game, false, "", "representation", "").Execute()
	if err != nil {
		http.Error(w, "Failed to create game", 500)
		log.Println("Supabase insert error:", err)
		return
	}

	var created []map[string]interface{}
	if err := json.Unmarshal(resp, &created); err != nil {
		http.Error(w, "Failed to decode game", 500)
		return
	}
	gameID := created[0]["id"].(string)

	cards := []map[string]interface{}{
		{"game_id": gameID, "title": "Feature A", "card_column": "Backlog"},
		{"game_id": gameID, "title": "Fix Bug", "card_column": "Dev"},
		{"game_id": gameID, "title": "Write Tests", "card_column": "Test"},
	}

	log.Printf("Inserting cards: %+v", cards)

	_, _, err = supabase.
		From("cards").
		Insert(cards, false, "", "representation", "").
		Execute()
	if err != nil {
		log.Println("Card insert error:", err)
	}

	log.Println("Card insert complete or error above")

	w.Write(resp)
}

func handleGetGame(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	id := strings.TrimPrefix(r.URL.Path, "/game/")
	if id == "" {
		http.Error(w, "Missing game ID", http.StatusBadRequest)
		return
	}

	resp, _, err := supabase.
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

func cardRouter(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// log.Println("Incoming path:", r.Method, r.URL.Path)

	path := strings.TrimPrefix(r.URL.Path, "/cards/")
	parts := strings.Split(path, "/")

	if r.Method == "GET" && len(parts) == 1 {
		// log.Println("Matched GET cards for game:", parts[0])
		handleGetCards(w, r, parts[0])
		return
	}

	if r.Method == "POST" && len(parts) == 2 && parts[1] == "move" {
		// log.Println("Matched POST move for card:", parts[0])
		handleMoveCard(w, r, parts[0])
		return
	}

	log.Println("No match in cardRouter")
	http.NotFound(w, r)
}

func handleGetCards(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	gameID = strings.TrimSpace(gameID)
	//REM: log.Println("Fetching cards for gameID:", gameID)

	if gameID == "" {
		http.Error(w, "Missing game ID", http.StatusBadRequest)
		return
	}

	resp, _, err := supabase.
		From("cards").
		Select("*", "exact", false).
		// Filter("game_id", "eq", gameID).
		Execute()

	// log.Println("All cards raw response:", string(resp))

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

	_, _, err := supabase.
		From("cards").
		Update(map[string]interface{}{"card_column": input.NewColumn}, "", "").
		Eq("id", cardID).
		Execute()
	// log.Printf("Moved card %s to %s", cardID, input.NewColumn)
	if err != nil {
		http.Error(w, "Failed to move card", 500)
		log.Println("Move card error:", err)
		return
	}
	// log.Println("handleMoveCard triggered")
	w.Write([]byte(`{"status":"ok"}`))
}

func gameRouter(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/game/")
	parts := strings.Split(path, "/")

	// POST /game/{id}/next-day
	if r.Method == "POST" && len(parts) == 2 && parts[1] == "next-day" {
		handleNextDay(w, r, parts[0])
		return
	}

	// fallback to existing GET logic
	if r.Method == "GET" && len(parts) == 1 {
		handleGetGame(w, r)
		return
	}

	if r.Method == "POST" && len(parts) == 2 && parts[1] == "join" {
		handleJoinGame(w, r, parts[0])
		return
	}

	if r.Method == "GET" && len(parts) == 2 && parts[1] == "players" {
		handleGetPlayers(w, r, parts[0])
		return
	}

	if r.Method == "POST" && len(parts) == 2 && parts[1] == "reset" {
		handleResetGame(w, r, parts[0])
		return
	}

	if r.Method == "POST" && len(parts) == 2 && parts[1] == "leave" {
		handleLeaveGame(w, r, parts[0])
		return
	}

	if r.Method == "POST" && len(parts) == 2 && parts[1] == "end" {
		handleEndGame(w, r, parts[0])
		return
	}

	http.NotFound(w, r)
}

func handleNextDay(w http.ResponseWriter, r *http.Request, gameID string) {
	// Fetch current game
	resp, _, err := supabase.
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

	// Update day
	updated, _, err := supabase.
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

func handleJoinGame(w http.ResponseWriter, r *http.Request, gameID string) {
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

	resp, _, err := supabase.
		From("players").
		Insert(player, false, "", "representation", "").
		Execute()

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Name already taken in this game", http.StatusConflict)
			return
		}
		http.Error(w, "Failed to join game", 500)
		log.Println("Join game error:", err)
		return
	}

	w.Write(resp)
}

func handleGetPlayers(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	resp, _, err := supabase.
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

func handleResetGame(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	// Load board config
	config, err := internal.LoadBoardConfig("config/board_config.json")
	if err != nil {
		http.Error(w, "Failed to load board config", 500)
		log.Println("Board config load error:", err)
		return
	}

	// 1. Delete all players
	_, _, err = supabase.
		From("players").
		Delete("", "").
		Eq("game_id", gameID).
		Execute()
	if err != nil {
		http.Error(w, "Failed to delete players", 500)
		log.Println("Delete players error:", err)
		return
	}

	// 2. Delete all cards
	_, _, err = supabase.
		From("cards").
		Delete("", "").
		Eq("game_id", gameID).
		Execute()
	if err != nil {
		http.Error(w, "Failed to delete cards", 500)
		log.Println("Delete cards error:", err)
		return
	}

	// 3. Update game columns in DB
	_, _, err = supabase.
		From("games").
		Update(map[string]interface{}{
			"columns": config.Columns,
			"day":     1,
		}, "", "").
		Eq("id", gameID).
		Execute()
	if err != nil {
		http.Error(w, "Failed to update game columns", 500)
		log.Println("Update columns error:", err)
		return
	}

	// 4. Insert initial cards from config
	var cardPayloads []map[string]interface{}
	for _, card := range config.InitialCards {
		cardPayload := map[string]interface{}{
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
		cardPayloads = append(cardPayloads, cardPayload)
	}

	_, _, err = supabase.From("cards").Insert(cardPayloads, false, "", "representation", "").Execute()
	if err != nil {
		http.Error(w, "Failed to insert reset cards", 500)
		log.Println("Insert reset cards error:", err)
		return
	}

	w.Write([]byte(`{"status":"reset complete"}`))
}

func handleLeaveGame(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	var input struct {
		PlayerID string `json:"player_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.PlayerID == "" {
		http.Error(w, "Missing player ID", 400)
		return
	}

	_, _, err := supabase.
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

func handleEndGame(w http.ResponseWriter, r *http.Request, gameID string) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	// Delete players
	_, _, err := supabase.From("players").Delete("", "").Eq("game_id", gameID).Execute()
	if err != nil {
		http.Error(w, "Failed to delete players", 500)
		log.Println("Delete players error:", err)
		return
	}

	// Delete cards
	_, _, err = supabase.From("cards").Delete("", "").Eq("game_id", gameID).Execute()
	if err != nil {
		http.Error(w, "Failed to delete cards", 500)
		log.Println("Delete cards error:", err)
		return
	}

	// Notify other players by inserting a "game ended" event
	_, _, err = supabase.
		From("game_events").
		Insert([]map[string]interface{}{
			{
				"game_id": gameID,
				"type":    "ended",
			},
		}, false, "", "representation", "").
		Execute()
	if err != nil {
		log.Println("Realtime event insert error:", err)
	}

	// Delete game
	_, _, err = supabase.From("games").Delete("", "").Eq("id", gameID).Execute()
	if err != nil {
		http.Error(w, "Failed to delete game", 500)
		log.Println("Delete game error:", err)
		return
	}

	w.Write([]byte(`{"status":"ended"}`))
}
