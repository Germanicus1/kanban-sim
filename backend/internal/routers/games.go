package routers

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal/handlers"
)

func InitGameRoutes() {
	http.HandleFunc("/games", handlers.CreateGame)        // POST /games
	http.HandleFunc("/games/get", handlers.GetGame)       // GET /games?id={id}
	http.HandleFunc("/games/board", handlers.GetBoard)    // GET /board?game_id={id}
	http.HandleFunc("/games/update", handlers.UpdateGame) // PUT /games?id={id}
	http.HandleFunc("/games/delete", handlers.DeleteGame) // DELETE /games?id={id}
	http.HandleFunc("/games/events", handlers.GetEvents)
}
