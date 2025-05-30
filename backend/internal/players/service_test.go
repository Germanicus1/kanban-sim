package players_test

import (
	"context"
	"testing"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/Germanicus1/kanban-sim/backend/internal/players"
	"github.com/google/uuid"
)

type mockRepo struct {
	wantID        uuid.UUID
	wantErr       error
	gotPlayer     models.Player
	wantPlayer    models.Player
	wantDeleteErr error
	gotDeleteID   uuid.UUID
}

func (m *mockRepo) CreatePlayer(ctx context.Context, cfg models.Player) (uuid.UUID, error) {
	m.gotPlayer = cfg
	return m.wantID, m.wantErr
}

func (m *mockRepo) GetPlayerByID(ctx context.Context, id uuid.UUID) (*models.Player, error) {
	m.gotPlayer = models.Player{ID: id, Name: m.wantPlayer.Name}
	return &m.gotPlayer, m.wantErr
}

func (m *mockRepo) UpdatePlayer(ctx context.Context, id uuid.UUID, name string) error {
	m.wantPlayer.Name = name
	return m.wantErr
}

func (m *mockRepo) DeletePlayer(ctx context.Context, id uuid.UUID) error {
	m.gotDeleteID = id
	return m.wantDeleteErr
}
func (m *mockRepo) ListPlayers(ctx context.Context, gameID uuid.UUID) ([]*models.Player, error) {
	if m.wantErr != nil {
		return nil, m.wantErr
	}
	return []*models.Player{&m.wantPlayer}, nil
}

func TestService_CreatePlayer(t *testing.T) {
	wantID := uuid.New()
	wantPlayer := models.Player{ID: wantID, Name: "Test Player"}
	mr := &mockRepo{wantID: wantID, wantPlayer: wantPlayer}
	svc := players.NewService(mr)

	gotID, err := svc.CreatePlayer(context.Background(), wantPlayer)
	if err != nil {
		t.Fatalf("CreatePlayer returned error: %v", err)
	}
	if gotID != wantID {
		t.Errorf("CreatePlayer got ID %v, want %v", gotID, wantID)
	}
	if mr.gotPlayer != wantPlayer {
		t.Errorf("CreatePlayer got player %v, want %v", mr.gotPlayer, wantPlayer)
	}
}

func TestService_GetPlayerByID(t *testing.T) {
	wantID := uuid.New()
	wantPlayer := models.Player{ID: wantID, Name: "Test Player"}
	mr := &mockRepo{wantPlayer: wantPlayer}
	svc := players.NewService(mr)

	gotPlayer, err := svc.GetPlayerByID(context.Background(), wantID)
	if err != nil {
		t.Fatalf("GetPlayer returned error: %v", err)
	}
	if gotPlayer.ID != wantID {
		t.Errorf("GetPlayer ID = %v; want %v", gotPlayer.ID, wantID)
	}
	if mr.gotPlayer.ID != wantID {
		t.Errorf("repo got id = %v; want %v", mr.gotPlayer.ID, wantID)
	}
}

func TestService_UpdatePlayer(t *testing.T) {
	wantID := uuid.New()
	wantName := "Updated Player"
	mr := &mockRepo{wantID: wantID, wantPlayer: models.Player{ID: wantID, Name: "Old Name"}}
	svc := players.NewService(mr)

	err := svc.UpdatePlayer(context.Background(), wantID, wantName)
	if err != nil {
		t.Fatalf("UpdatePlayer returned error: %v", err)
	}
	if mr.wantPlayer.Name != wantName {
		t.Errorf("UpdatePlayer name = %v; want %v", mr.wantPlayer.Name, wantName)
	}
}

func TestService_DeletePlayer(t *testing.T) {
	wantID := uuid.New()
	mr := &mockRepo{wantDeleteErr: nil}
	svc := players.NewService(mr)

	err := svc.DeletePlayer(context.Background(), wantID)
	if err != nil {
		t.Fatalf("DeletePlayer returned error: %v", err)
	}
	if mr.gotDeleteID != wantID {
		t.Errorf("repo got id = %v; want %v", mr.gotDeleteID, wantID)
	}
}

func TestService_ListPlayers(t *testing.T) {
	wantID := uuid.New()
	wantPlayer := models.Player{ID: wantID, Name: "Test Player"}
	mr := &mockRepo{wantPlayer: wantPlayer}
	svc := players.NewService(mr)

	gotPlayers, err := svc.ListPlayers(context.Background())
	if err != nil {
		t.Fatalf("ListPlayers returned error: %v", err)
	}
	if len(gotPlayers) != 1 {
		t.Fatalf("ListPlayers got %d players, want 1", len(gotPlayers))
	}
	if gotPlayers[0].ID != wantID {
		t.Errorf("ListPlayers got player ID %v, want %v", gotPlayers[0].ID, wantID)
	}
	if gotPlayers[0].Name != wantPlayer.Name {
		t.Errorf("ListPlayers got player name %v, want %v", gotPlayers[0].Name, wantPlayer.Name)
	}
}
