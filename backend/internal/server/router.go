// internal/server/router.go
package server

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal/handlers"
)

func NewRouter(ah *handlers.AppHandler, gh *handlers.GameHandler) http.Handler {
	mux := http.NewServeMux()

	// public pages
	mux.HandleFunc("GET /", ah.Home)
	mux.HandleFunc("GET /ping", ah.Ping)
	mux.HandleFunc("GET /openapi.yaml", ah.OpenAPI)
	mux.HandleFunc("GET /docs", ah.DocsRedirect)
	mux.Handle(
		"GET /docs/",
		http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))),
	)

	// games API
	mux.HandleFunc("POST /games", gh.CreateGame)
	mux.HandleFunc("GET /games/{id}", gh.GetGame)
	mux.HandleFunc("GET /games/{id}/board", gh.GetBoard)
	mux.HandleFunc("PATCH /games/{id}", gh.UpdateGame)
	mux.HandleFunc("DELETE /games/{id}", gh.DeleteGame)

	return mux
}
