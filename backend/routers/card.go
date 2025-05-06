package routers

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/handlers"
)

func RegisterCardRoutes() {
	http.HandleFunc("/cards/", handlers.CardRouter)
}
