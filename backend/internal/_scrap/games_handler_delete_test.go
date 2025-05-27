package handlers

import (
	"context"
	"testing"

	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

type fakeService struct {
	// Mocked methods for testing
	wantID  uuid.UUID
	wantErr error
	called  bool
}

func (f *fakeService) CreateGame(ctx context.Context, cfg models.BoardConfig) (uuid.UUID, error) {
	panic("unused")
}
func (f *fakeService) GetBoard(ctx context.Context, id uuid.UUID) (models.Board, error) {
	panic("unused")
}
func (f *fakeService) GetGame(ctx context.Context, id uuid.UUID) (models.Game, error) {
	panic("unused")
}

func (f *fakeService) DeleteGame(ctx context.Context, id uuid.UUID) error {
	f.called = true
	return f.wantErr
}

func TestService_DeleteGame_Success(t *testing.T) {
	// Setup
	wantID := uuid.New()
	fakeSvc := &fakeService{wantErr: nil}
	h := NewGameHandler(fakeSvc)

	// Call the method
	err := h.DeleteGame(context.Background(), wantID)

	// Assertions
	if err != nil {
		t.Fatalf("DeleteGame returned error: %v", err)
	}
	if !fakeSvc.called {
		t.Fatal("DeleteGame was not called on the service")
	}
}
