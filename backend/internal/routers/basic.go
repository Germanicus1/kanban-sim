package routers

import (
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal/handlers"
)

func InitRoutes() {
	http.HandleFunc("/", handlers.HandleHome)
	http.HandleFunc("/ping", handlers.HandlePing)
	// http.HandleFunc("/create-game", handlers.HandleCreateGame)
	// http.HandleFunc("/cards/", handlers.CardRouter)
	// http.HandleFunc("/game/", handlers.GameRouter)
	// Serve the spec and docs
	// http.Handle("/openapi.yaml", http.FileServer(http.Dir("./")))
	// Serve the docs folder under /docs/*

	// redirect /docs → /docs/
	// http.Handle("/docs", http.RedirectHandler("/docs/", http.StatusMovedPermanently))

	// // serve everything under ./docs at /docs/*
	// http.Handle(
	// 	"/docs/",
	// 	http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))),
	// )

}
