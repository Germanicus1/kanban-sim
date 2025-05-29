package database

import (
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func setupEnv(t *testing.T) {
	// Construct the path to the .env file one level up
	envPath := filepath.Join("..", "..", ".env")

	// Load environment variables from .env
	if err := godotenv.Load(envPath); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
}

func TestInitDB(t *testing.T) {
	setupEnv(t)

	// Initialize the database
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing DB: %v", err)
		}
	}()

	// Check the connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping DB: %v", err)
	}
}

func TestMigrate(t *testing.T) {
	setupEnv(t)

	// Initialize the database
	db, err := InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing DB: %v", err)
		}
	}()

	// Define migration directory
	migrationDir := "./migrations"

	// Run migrations
	err = Migrate(db, migrationDir)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}
}
