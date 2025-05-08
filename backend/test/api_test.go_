package test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:8080"

var createdGameID string

func TestCreateGame(t *testing.T) {
	resp, err := http.Post(baseURL+"/create-game", "application/json", nil)
	if err != nil {
		t.Fatalf("Create game request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var body []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	id, ok := body[0]["id"].(string)
	if !ok || id == "" {
		t.Fatal("Missing game ID in response")
	}
	createdGameID = id
	t.Logf("✅ Created game: %s", createdGameID)
}

func TestJoinGame(t *testing.T) {
	if createdGameID == "" {
		t.Skip("Game not created")
	}
	payload := map[string]string{"name": "TestPlayer"}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(baseURL+"/game/"+createdGameID+"/join", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Join game failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}
	t.Log("✅ Joined game")
}

func TestResetGame(t *testing.T) {
	if createdGameID == "" {
		t.Skip("Game not created")
	}
	resp, err := http.Post(baseURL+"/game/"+createdGameID+"/reset", "application/json", nil)
	if err != nil {
		t.Fatalf("Reset game failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("reset complete")) {
		t.Fatalf("Unexpected reset response: %s", string(body))
	}
	t.Log("✅ Reset game")
}

func TestGetCards(t *testing.T) {
	if createdGameID == "" {
		t.Skip("Game not created")
	}
	url := baseURL + "/cards/" + createdGameID
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to get cards: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("Expected 200, got %d", resp.StatusCode)
	}

	var cards []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil {
		t.Fatalf("Failed to decode cards JSON: %v", err)
	}

	if len(cards) == 0 {
		t.Error("Expected at least 1 card after game creation")
	} else {
		t.Logf("✅ Fetched %d cards", len(cards))
	}
}

func TestMoveCard(t *testing.T) {
	if createdGameID == "" {
		t.Skip("Game not created")
	}

	// Fetch cards
	resp, err := http.Get(baseURL + "/cards/" + createdGameID)
	if err != nil {
		t.Fatalf("Failed to get cards: %v", err)
	}
	defer resp.Body.Close()

	var cards []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&cards); err != nil || len(cards) == 0 {
		t.Fatal("No cards found to move")
	}

	cardID := cards[0]["id"].(string)

	payload := map[string]string{"new_column": "test"}
	body, _ := json.Marshal(payload)

	moveResp, err := http.Post(baseURL+"/cards/"+cardID+"/move", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to move card: %v", err)
	}
	defer moveResp.Body.Close()

	if moveResp.StatusCode != 200 {
		t.Fatalf("Expected 200 on move, got %d", moveResp.StatusCode)
	}
	t.Logf("✅ Moved card %s to 'test'", cardID)
}

func TestEndGame(t *testing.T) {
	if createdGameID == "" {
		t.Skip("Game not created")
	}
	resp, err := http.Post(baseURL+"/game/"+createdGameID+"/end", "application/json", nil)
	if err != nil {
		t.Fatalf("End game failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("ended")) {
		t.Fatalf("Unexpected end response: %s", string(body))
	}
	t.Log("✅ Ended game")
}
