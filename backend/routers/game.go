package routers

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/handlers"
)

func InitRoutes() {
	http.HandleFunc("/ping", handlers.HandlePing)
	http.HandleFunc("/create-game", handlers.HandleCreateGame)
	http.HandleFunc("/cards/", handlers.CardRouter)
	http.HandleFunc("/game/", handlers.GameRouter)
}
