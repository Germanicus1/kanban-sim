// internal/games/service_test.go
package games

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/Germanicus1/kanban-sim/internal/response"
	"github.com/google/uuid"
)

// mockRepo implements only the bits of Repository we need.
type mockRepo struct {
	wantID        uuid.UUID
	wantErr       error
	gotCfg        models.BoardConfig
	gotGame       uuid.UUID
	wantBoard     models.Board
	wantDeleteID  uuid.UUID
	wantDeleteErr error
	gotDeleteID   uuid.UUID
}

func (m *mockRepo) CreateGame(ctx context.Context, cfg models.BoardConfig) (uuid.UUID, error) {
	m.gotCfg = cfg
	return m.wantID, m.wantErr
}
func (m *mockRepo) GetBoard(ctx context.Context, id uuid.UUID) (models.Board, error) {
	m.gotGame = id
	return m.wantBoard, m.wantErr
}
func (m *mockRepo) GetGameByID(ctx context.Context, id uuid.UUID) (models.Game, error) {
	return models.Game{}, nil
}

func (m *mockRepo) DeleteGame(ctx context.Context, id uuid.UUID) error {
	m.gotDeleteID = id
	return m.wantDeleteErr
}

func (m *mockRepo) UpdateGame(ctx context.Context, id uuid.UUID, day int) error {
	m.wantID = id
	return nil
}

func TestService_CreateGame(t *testing.T) {
	wantID := uuid.New()
	cfg := models.BoardConfig{} // you can fill with dummy data
	mr := &mockRepo{wantID: wantID}
	svc := NewService(mr)

	gotID, err := svc.CreateGame(context.Background(), cfg)
	if err != nil {
		t.Fatalf("CreateGame returned error: %v", err)
	}
	if gotID != wantID {
		t.Errorf("CreateGame = %v; want %v", gotID, wantID)
	}
	if !reflect.DeepEqual(mr.gotCfg, cfg) {
		t.Errorf("repo got cfg = %+v; want %+v", mr.gotCfg, cfg)
	}
}

func TestService_CreateGame_Error(t *testing.T) {
	cfg := models.BoardConfig{}
	wantErr := errors.New("boom")
	mr := &mockRepo{wantErr: wantErr}
	svc := NewService(mr)

	if _, err := svc.CreateGame(context.Background(), cfg); err != wantErr {
		t.Errorf("CreateGame error = %v; want %v", err, wantErr)
	}
}

func TestService_GetBoard(t *testing.T) {
	wantID := uuid.New()
	wantBoard := models.Board{GameID: wantID}
	mr := &mockRepo{wantBoard: wantBoard}
	svc := NewService(mr)

	got, err := svc.GetBoard(context.Background(), wantID)
	if err != nil {
		t.Fatalf("GetBoard returned error: %v", err)
	}
	if got.GameID != wantBoard.GameID {
		t.Errorf("GetBoard.GameID = %v; want %v", got.GameID, wantBoard.GameID)
	}
	if mr.gotGame != wantID {
		t.Errorf("repo got id = %v; want %v", mr.gotGame, wantID)
	}
}

func TestService_DeleteGame_Success(t *testing.T) {
	id := uuid.New()
	mr := &mockRepo{wantDeleteErr: nil}
	svc := NewService(mr)

	err := svc.DeleteGame(context.Background(), id)
	if err != nil {
		t.Fatalf("DeleteGame returned error: %v", err)
	}
	if mr.gotDeleteID != id {
		t.Errorf("repo.DeleteGame called with %v; want %v", mr.gotDeleteID, id)
	}
}

func TestService_DeleteGame_NotFound(t *testing.T) {
	id := uuid.New()
	mr := &mockRepo{wantDeleteErr: response.ErrNotFound}
	svc := NewService(mr)

	err := svc.DeleteGame(context.Background(), id)
	if !errors.Is(err, response.ErrNotFound) {
		t.Errorf("DeleteGame error = %v; want ErrNotFound", err)
	}
}

func TestService_UpdateGame_Success(t *testing.T) {
	id := uuid.New()
	day := 1
	mr := &mockRepo{}
	svc := NewService(mr)

	err := svc.UpdateGame(context.Background(), id, day)
	if err != nil {
		t.Fatalf("UpdateGame returned error: %v", err)
	}
	if mr.wantID != id {
		t.Errorf("repo.UpdateGame called with %v; want %v", mr.wantID, id)
	}
}
