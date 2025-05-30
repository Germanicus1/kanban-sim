package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

// fakeService implements games.ServiceInterface for testing DeleteGame.
type fakeService struct {
	calledID uuid.UUID
	retErr   error
}

func (f *fakeService) CreateGame(ctx context.Context, cfg models.BoardConfig) (uuid.UUID, error) {
	panic("unused")
}
func (f *fakeService) GetBoard(ctx context.Context, id uuid.UUID) (models.Board, error) {
	panic("unused")
}
func (f *fakeService) GetGame(ctx context.Context, id uuid.UUID) (models.Game, error) {
	f.calledID = id
	return models.Game{}, f.retErr
}
func (f *fakeService) DeleteGame(ctx context.Context, id uuid.UUID) error {
	f.calledID = id
	return f.retErr
}

func (f *fakeService) UpdateGame(ctx context.Context, id uuid.UUID, day int) error {
	f.calledID = id
	return f.retErr
}

func TestGameHandler_GetGame_Success(t *testing.T) {
	svc := &fakeService{retErr: nil}
	h := NewGameHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /games/{id}", h.GetGame)

	id := uuid.New()
	req := httptest.NewRequest("GET", "/games/"+id.String(), nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}
	if svc.calledID != id {
		t.Errorf("service.GetGame called with %v; want %v", svc.calledID, id)
	}
}

func TestGameHandler_UpdateGame_Success(t *testing.T) {
	svc := &fakeService{retErr: nil}
	h := NewGameHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /games/{id}", h.UpdateGame)

	id := uuid.New()
	// cfg := models.Game{Day: 1}
	body := `{"day": 1}`
	req := httptest.NewRequest("PATCH", "/games/"+id.String(), strings.NewReader(body))
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusNoContent)
	}
	if svc.calledID != id {
		t.Errorf("service.UpdateGame called with %v; want %v", svc.calledID, id)
	}
}

func TestGameHandler_DeleteGame_Success(t *testing.T) {
	svc := &fakeService{retErr: nil}
	h := NewGameHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /games/{id}", h.DeleteGame)

	id := uuid.New()
	req := httptest.NewRequest("DELETE", "/games/"+id.String(), nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusNoContent)
	}
	if svc.calledID != id {
		t.Errorf("service.DeleteGame called with %v; want %v", svc.calledID, id)
	}
}

func TestGameHandler_DeleteGame_NotFound(t *testing.T) {
	svc := &fakeService{retErr: response.ErrNotFound}
	h := NewGameHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /games/{id}", h.DeleteGame)

	id := uuid.New()
	req := httptest.NewRequest("DELETE", "/games/"+id.String(), nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusNotFound)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"success":false`) || !strings.Contains(body, response.ErrGameNotFound) {
		t.Errorf("body = %q; want JSON error envelope with game_not_found", body)
	}
}

func TestGameHandler_DeleteGame_BadID(t *testing.T) {
	svc := &fakeService{retErr: nil}
	h := NewGameHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /games/{id}", h.DeleteGame)

	req := httptest.NewRequest("DELETE", "/games/not-a-uuid", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rr.Body.String(), `"success":false`) {
		t.Errorf("body = %q; want JSON error envelope", rr.Body.String())
	}
}
