package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/Germanicus1/kanban-sim/routers"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")

	var err error
	internal.Supabase, err = internal.InitSupabase(supabaseURL, supabaseKey)
	if err != nil {
		log.Fatal("Failed to initialize Supabase:", err)
	}

	routers.InitRoutes()

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
