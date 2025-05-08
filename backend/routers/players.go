package routers

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/handlers"
)

func InitPlayerRoutes() {
	http.HandleFunc("/players", handlers.CreatePlayer)        // POST /players
	http.HandleFunc("/players/get", handlers.GetPlayer)       // GET /players?id={id}
	http.HandleFunc("/players/update", handlers.UpdatePlayer) // PUT /players?id={id}
	http.HandleFunc("/players/delete", handlers.DeletePlayer) // DELETE /players?id={id}
}
