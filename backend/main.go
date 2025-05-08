package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/Germanicus1/kanban-sim/routers"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	migrateOnly := flag.Bool("migrate-only", false, "Run migrations only")
	flag.Parse()

	db, err := internal.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	if *migrateOnly {
		log.Println("Running migrations...")
		err = internal.Migrate(db, "./migrations")
		if err != nil {
			log.Fatal("Failed to migrate DB:", err)
		}
		return
	}

	err = internal.Migrate(db, "./migrations")
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

	routers.InitGameRoutes()

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
