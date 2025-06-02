package columns_test

import (
	"context"
	"testing"

	"github.com/Germanicus1/kanban-sim/backend/internal/columns"
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type mockRepo struct {
	wantID     uuid.UUID
	wantErr    error
	gotColumn  models.Column
	wantColumn models.Column
}

func (m *mockRepo) GetColumnsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Column, error) {
	if m.wantErr != nil {
		return nil, m.wantErr
	}
	return []models.Column{m.wantColumn}, nil
}

func TestService_GetColumnsByGameID(t *testing.T) {
	wantID := uuid.New()
	wantColumn := models.Column{ID: wantID, Title: "Test Column", OrderIndex: 0, WIPLimit: nil, Type: "active"}
	mr := &mockRepo{wantID: wantID, wantColumn: wantColumn}
	svc := columns.NewService(mr)

	gotColumns, err := svc.GetColumnsByGameID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetColumnsByGameID returned error: %v", err)
	}
	if len(gotColumns) != 1 || gotColumns[0].ID != wantColumn.ID || gotColumns[0].Title != wantColumn.Title || gotColumns[0].OrderIndex != wantColumn.OrderIndex || gotColumns[0].Type != wantColumn.Type {
		t.Errorf("GetColumnsByGameID got %v, want %v", gotColumns, []models.Column{wantColumn})
	}
}
