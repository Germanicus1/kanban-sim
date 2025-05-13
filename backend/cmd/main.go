package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Germanicus1/kanban-sim/internal/database"
	"github.com/Germanicus1/kanban-sim/internal/routers"
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

	routers.InitRoutes()
	routers.InitGameRoutes()
	// routers.InitCardRoutes()
	// routers.InitPlayerRoutes()

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
