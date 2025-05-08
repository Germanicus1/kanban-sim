package routers

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/handlers"
)

func InitCardRoutes() {
	http.HandleFunc("/cards", handlers.CreateGame)        // POST /games
	http.HandleFunc("/cards/get", handlers.GetGame)       // GET /games?id={id}
	http.HandleFunc("/cards/update", handlers.UpdateGame) // PUT /games?id={id}
	http.HandleFunc("/cards/delete", handlers.DeleteGame) // DELETE /games?id={id}
}
