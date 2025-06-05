package server

import (
	"net/http"

	_ "github.com/Germanicus1/kanban-sim/backend/apidocs"
	"github.com/Germanicus1/kanban-sim/backend/internal/handlers"
	"github.com/Germanicus1/kanban-sim/backend/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(
	ah *handlers.AppHandler,
	gh *handlers.GameHandler,
	ph *handlers.PlayerHandler,
	ch *handlers.ColumnsHandler,
) http.Handler {
	mainMux := http.NewServeMux()

	// public pages
	mainMux.HandleFunc("GET /", ah.Home)
	mainMux.HandleFunc("GET /ping", ah.Ping)
	mainMux.HandleFunc("GET /openapi.yaml", ah.OpenAPI)
	mainMux.Handle("GET /apidocs/", httpSwagger.WrapHandler)

	privateMux := http.NewServeMux()
	// games API
	privateMux.HandleFunc("POST /games", gh.CreateGame)
	privateMux.HandleFunc("GET /games", gh.ListGames)
	privateMux.HandleFunc("GET /games/{id}", gh.GetGame)
	privateMux.HandleFunc("GET /games/{id}/board", gh.GetBoard)
	privateMux.HandleFunc("PATCH /games/{id}", gh.UpdateGame)
	privateMux.HandleFunc("DELETE /games/{id}", gh.DeleteGame)

	// players API
	privateMux.HandleFunc("POST /players", ph.CreatePlayer)
	privateMux.HandleFunc("GET /players/{id}", ph.GetPlayerByID)
	privateMux.HandleFunc("PATCH /players/{id}", ph.UpdatePlayer)
	privateMux.HandleFunc("DELETE /players", ph.DeletePlayer)
	privateMux.HandleFunc("GET /games/{game_id}/players", ph.ListPlayersByGameID)

	// columns API
	privateMux.HandleFunc("GET /games/{id}/columns", ch.GetColumnsByGameID)

	mainMux.Handle("/", middleware.APIKeyAuth(privateMux))

	return mainMux
}
