package main

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func setupEnv(t *testing.T) {
	envPath := filepath.Join(".", ".env")
	if err := godotenv.Load(envPath); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}
}

func runMigration(t *testing.T) {
	cmd := exec.Command("./main", "--migrate-only")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}
}

func startServer(t *testing.T) *exec.Cmd {
	cmd := exec.Command("./main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Allow server time to start
	time.Sleep(2 * time.Second)

	return cmd
}

func stopServer(cmd *exec.Cmd) {
	if cmd != nil {
		cmd.Process.Kill()
	}
}

func TestMainApp(t *testing.T) {
	setupEnv(t)
	runMigration(t)

	server := startServer(t)
	defer stopServer(server)

	// Send a request to the root route
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}
}
