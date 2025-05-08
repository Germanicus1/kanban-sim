package routers

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/handlers"
)

func InitCardRoutes() {
	http.HandleFunc("/cards", handlers.CreateCard)        // POST /cards
	http.HandleFunc("/cards/get", handlers.GetCard)       // GET /cards?id={id}
	http.HandleFunc("/cards/update", handlers.UpdateCard) // PUT /cards?id={id}
	http.HandleFunc("/cards/delete", handlers.DeleteCard) // DELETE /cards?id={id}
}
