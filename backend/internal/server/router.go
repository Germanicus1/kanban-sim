package server

import (
	"net/http"

	_ "github.com/Germanicus1/kanban-sim/backend/apidocs"
	"github.com/Germanicus1/kanban-sim/backend/internal/handlers"
	"github.com/Germanicus1/kanban-sim/backend/internal/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type route struct {
	Pattern string
	Handler http.HandlerFunc
}

func NewRouter(
	ah *handlers.AppHandler,
	gh *handlers.GameHandler,
	ph *handlers.PlayerHandler,
	ch *handlers.ColumnsHandler,
) (mux *http.ServeMux) {
	// public pages
	mux = http.NewServeMux()

	// ─── PRIVATE ROUTES (wrapped in APIKeyAuth) ─────────────────────────────────
	privateRoutes := []route{
		{"POST /games", gh.CreateGame},
		{"GET /games", gh.ListGames},
		{"GET /games/{id}", gh.GetGame},
		{"GET /games/{id}/board", gh.GetBoard},
		{"PATCH /games/{id}", gh.UpdateGame},
		{"DELETE /games/{id}", gh.DeleteGame},

		{"POST /players", ph.CreatePlayer},
		{"GET /players/{id}", ph.GetPlayerByID},
		{"PATCH /players/{id}", ph.UpdatePlayer},
		{"DELETE /players", ph.DeletePlayer},
		{"GET /games/{game_id}/players", ph.ListPlayersByGameID},

		{"GET /games/{id}/columns", ch.GetColumnsByGameID},
	}

	for _, r := range privateRoutes {
		// wrap each handler func in APIKeyAuth, then register with mux.Handle
		// (http.HandlerFunc already implements http.Handler)
		mux.Handle(r.Pattern, middleware.APIKeyAuth(http.HandlerFunc(r.Handler)))
	}

	// ─── PUBLIC ROUTES ────────────────────────────────────────────────────────
	mux.HandleFunc("GET /", ah.Home)
	mux.HandleFunc("GET /ping", ah.Ping)
	mux.HandleFunc("GET /openapi.yaml", ah.OpenAPI)
	mux.Handle("GET /apidocs/", httpSwagger.WrapHandler)

	return mux
}
