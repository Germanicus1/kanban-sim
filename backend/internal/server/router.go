package server

import (
	"net/http"

	_ "github.com/Germanicus1/kanban-sim/backend/apidocs"
	"github.com/Germanicus1/kanban-sim/backend/internal/handlers"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(ah *handlers.AppHandler, gh *handlers.GameHandler, ph *handlers.PlayerHandler, ch *handlers.ColumnsHandler) http.Handler {
	mux := http.NewServeMux()

	// public pages
	mux.HandleFunc("GET /", ah.Home)
	mux.HandleFunc("GET /ping", ah.Ping)
	mux.HandleFunc("GET /openapi.yaml", ah.OpenAPI)
	mux.Handle("GET /apidocs/", httpSwagger.WrapHandler)

	// games API
	mux.HandleFunc("POST /games", gh.CreateGame)
	mux.HandleFunc("GET /games", gh.ListGames)
	mux.HandleFunc("GET /games/{id}", gh.GetGame)
	mux.HandleFunc("GET /games/{id}/board", gh.GetBoard)
	mux.HandleFunc("PATCH /games/{id}", gh.UpdateGame)
	mux.HandleFunc("DELETE /games/{id}", gh.DeleteGame)

	// players API
	mux.HandleFunc("POST /players", ph.CreatePlayer)
	mux.HandleFunc("GET /players/{id}", ph.GetPlayerByID)
	mux.HandleFunc("PATCH /players/{id}", ph.UpdatePlayer)
	mux.HandleFunc("DELETE /players", ph.DeletePlayer)
	mux.HandleFunc("GET /games/{game_id}/players", ph.ListPlayersByGameID)

	// columns API
	mux.HandleFunc("GET /games/{id}/columns", ch.GetColumnsByGameID)

	return mux
}
