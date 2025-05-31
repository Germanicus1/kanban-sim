//go:generate swag init -g ./cmd/main.go -d ../ -o ../apidocs --md ../apidocs/md
// @title        Kanban-Sim API
// @version      1.0
// @description  A simple Kanban simulation API.
// @termsOfService http://example.com/terms/
// @contact.name  Peter Kerschbaumer
// @contact.email you@example.com
// @license.name MIT
// @license.url  https://opensource.org/licenses/MIT
// @host         localhost:8080
// @BasePath     /

package main

import (
	"context"
	"flag"

	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Germanicus1/kanban-sim/backend/internal/database"
	"github.com/Germanicus1/kanban-sim/backend/internal/games"
	"github.com/Germanicus1/kanban-sim/backend/internal/handlers"
	"github.com/Germanicus1/kanban-sim/backend/internal/players"
	"github.com/Germanicus1/kanban-sim/backend/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Failed to load .env: %v", err)
	}

	// Parse flags
	migrateOnly := flag.Bool("migrate-only", false, "Run migrations only")
	flag.Parse()

	// Initialize DB
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database: ", err)
	}

	// Migrations only
	if *migrateOnly {
		log.Println("Running migrations...")
		if err := database.Migrate(db, "./internal/database/migrations"); err != nil {
			log.Fatal("Failed to migrate DB: ", err)
		}
		return
	}

	// Auto-migrate on startup
	if err := database.Migrate(db, "./internal/database/migrations"); err != nil {
		log.Fatal("Failed to migrate DB: ", err)
	}

	// Setup services and handlers
	gameRepo := games.NewSQLRepo(db)
	playerRepo := players.NewSQLRepo(db)
	gameSvc := games.NewService(gameRepo)
	playerSvc := players.NewService(playerRepo)

	gh := handlers.NewGameHandler(gameSvc)
	ah := handlers.NewAppHandler()
	ph := handlers.NewPlayerHandler(playerSvc)
	router := server.NewRouter(ah, gh, ph)

	// Configure HTTP server with timeouts
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown on SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	// Start server
	log.Println("Server running at http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}
