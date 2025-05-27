package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/games"
	"github.com/Germanicus1/kanban-sim/internal/handlers"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Failed to load .env: %v", err)
	}

	// Parse command-line flags This is useful for running migrations without
	// starting the server. You can run the server with `go run main.go` or just
	// run migrations with `go run main.go -migrate-only`
	migrateOnly := flag.Bool("migrate-only", false, "Run migrations only")
	flag.Parse()

	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Run migrations if the flag is set. Server is not started in this case.
	if *migrateOnly {
		log.Println("Running migrations...")
		err = database.Migrate(db, "./internal/database/migrations")
		if err != nil {
			log.Fatal("Failed to migrate DB:", err)
		}
		return
	}

	err = database.Migrate(db, "./internal/database/migrations")
	if err != nil {
		log.Fatal("Failed to migrate DB:", err)
	}

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server...")
		os.Exit(0)
	}()

	mux := http.NewServeMux()

	repo := games.NewSQLRepo(db) // your sql_repo.go
	svc := games.NewService(repo)
	gh := handlers.NewGameHandler(svc) // game_handler.go
	appH := handlers.NewAppHandler()

	mux.HandleFunc("GET /", appH.Home)
	mux.HandleFunc("GET /ping", appH.Ping)
	mux.HandleFunc("POST /games", gh.CreateGame)
	mux.HandleFunc("GET /games/{id}", gh.GetGame)
	mux.HandleFunc("GET /games/{id}/board", gh.GetBoard)
	mux.HandleFunc("DELETE /games/{id}", gh.DeleteGame)
	mux.HandleFunc("PATCH /games/{id}", gh.UpdateGame)

	// API documentation
	mux.HandleFunc("GET /openapi.yaml", appH.OpenAPI)
	mux.HandleFunc("GET /docs", appH.DocsRedirect)
	mux.Handle(
		"GET /docs/",
		http.StripPrefix("/docs/", http.FileServer(http.Dir("./docs"))),
	)

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
