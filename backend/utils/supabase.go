package utils

import (
	"log"
	"os"

	supa "github.com/supabase-community/supabase-go"
)

var Client *supa.Client

func InitSupabase() {
	url := os.Getenv("SUPABASE_URL")
	key := os.Getenv("SUPABASE_KEY")

	client, err := supa.NewClient(url, key, nil)
	if err != nil {
		log.Fatalf("Failed to create Supabase client: %v", err)
	}
	Client = client
}
