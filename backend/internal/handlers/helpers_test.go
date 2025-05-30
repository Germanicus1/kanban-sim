package handlers_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Germanicus1/kanban-sim/backend/internal/database"
	"github.com/joho/godotenv"
)

var db *sql.DB

func setupEnv(t *testing.T) {
	envPath := filepath.Join("..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
}

func SetupDB(t *testing.T, tableName string) {
	setupEnv(t)
	var err error
	db, err = database.InitDB()
	if err != nil {
		t.Fatalf("Failed to initialize DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatalf("Database connection failed: %v", err)
	}

	_, err = db.Exec("DELETE FROM " + tableName)
	if err != nil {
		t.Fatalf("Failed to clean games table: %v", err)
	}
}

func TearDownDB() {
	if db != nil {
		db.Close()
	}
}
