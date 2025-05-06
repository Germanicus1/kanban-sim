package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal"
)

func HandleCreateGame(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	game := []map[string]interface{}{
		{
			"day":     1,
			"columns": []string{"Backlog", "Dev", "Test", "Done"},
		},
	}

	resp, _, err := internal.Supabase.From("games").Insert(game, false, "", "representation", "").Execute()
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

	_, _, err = internal.Supabase.
		From("cards").
		Insert(cards, false, "", "representation", "").
		Execute()
	if err != nil {
		log.Println("Card insert error:", err)
	}

	log.Println("Card insert complete or error above")

	w.Write(resp)
}
