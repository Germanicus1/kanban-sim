package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Germanicus1/kanban-sim/backend/internal"
	"github.com/Germanicus1/kanban-sim/backend/routers"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")
	var err error
	internal.Supabase, err = internal.InitSupabase(url, key)
	if err != nil {
		log.Fatal("Failed to connect to Supabase:", err)
	}

	routers.InitRoutes()
	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
