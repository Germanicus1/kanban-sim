package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Germanicus1/kanban-sim/backend/internal/handlers"
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

// fakeService implements players.ServiceInterface.
type fakeService struct {
	calledID uuid.UUID
	retErr   error
}

func (f *fakeService) CreatePlayer(ctx context.Context, gameID uuid.UUID, name string) (uuid.UUID, error) {
	f.calledID = gameID
	return uuid.New(), f.retErr
}

func (f *fakeService) DeletePlayer(ctx context.Context, id uuid.UUID) error {
	f.calledID = id
	return f.retErr
}

func (f *fakeService) GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
	f.calledID = id
	if f.retErr != nil {
		return nil, f.retErr
	}
	return &models.Player{ID: id, Name: "Test Player", GameID: uuid.New()}, nil
}
func (f *fakeService) UpdatePlayer(ctx context.Context, id uuid.UUID, name string) error {
	f.calledID = id
	return f.retErr
}
func (f *fakeService) ListPlayersByGameID(ctx context.Context, gameID uuid.UUID) ([]*models.Player, error) {
	f.calledID = gameID
	if f.retErr != nil {
		return nil, f.retErr
	}
	return []*models.Player{{ID: uuid.New(), Name: "Test Player", GameID: gameID}}, nil
}

func TestPlayerHandler_CreatePlayer(t *testing.T) {
	svc := &fakeService{retErr: nil}
	h := handlers.NewPlayerHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/players", h.CreatePlayer)

	// create a valid JSON payload
	gameID := uuid.New()
	body := fmt.Sprintf(`{"game_id": "%s", "name": "Test Player"}`, gameID.String())
	req := httptest.NewRequest("POST", "/players", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// 1) Status code
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}

	// 2) Read the entire body as a string
	respBytes, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	playerID := strings.TrimSpace(string(respBytes))
	if playerID == "" {
		t.Fatalf("expected player ID to be returned, got empty string")
	}

	// 3) Verify the service was called with the correct game ID
	if svc.calledID == uuid.Nil {
		t.Fatalf("expected service to be called with a valid game ID, got nil")
	}
	if svc.calledID != gameID {
		t.Errorf("expected service called with game ID %s, got %s", gameID, svc.calledID)
	}
}

func TestPlayerHandler_DeletePlayer(t *testing.T) {
	svc := &fakeService{retErr: nil}
	h := handlers.NewPlayerHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("/players/", h.DeletePlayer)
	// Create a request to delete a player with a specific ID
	playerID := uuid.New()
	reqBody := fmt.Sprintf(`{"id":"%s"}`, playerID)

	req := httptest.NewRequest("DELETE", "/players/", strings.NewReader(reqBody))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	// 1) Check the status code
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}
	// 2) Verify the service was called with the correct player ID
	if svc.calledID == uuid.Nil {
		t.Fatalf("expected service to be called with a valid player ID, got nil")
	}
	if svc.calledID != playerID {
		t.Errorf("expected service called with player ID %s, got %s", playerID, svc.calledID)
	}
}

func TestPlayerHandler_ListPlayersByGameID(t *testing.T) {
	svc := &fakeService{retErr: nil}
	h := handlers.NewPlayerHandler(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /games/{game_id}/players", h.ListPlayersByGameID)

	gameID := uuid.New()
	req := httptest.NewRequest("GET", "/games/"+gameID.String()+"/players", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// 1) Check the status code
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}

	// 2) Verify the service was called with the correct game ID
	if svc.calledID == uuid.Nil {
		t.Fatalf("expected service to be called with a valid game ID, got nil")
	}
	if svc.calledID != gameID {
		t.Errorf("expected service called with game ID %s, got %s", gameID, svc.calledID)
	}
}
