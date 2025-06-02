package columns_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Germanicus1/kanban-sim/backend/internal/columns"
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSQLRepo_GetColumnsByGameID(t *testing.T) {
	const query = `SELECT id, title, wip_limit, col_type, parent_id, order_index FROM columns WHERE game_id = $1 ORDER BY parent_id, order_index`

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock, id uuid.UUID)
		gameID    uuid.UUID
		expected  []models.Column
	}{
		{
			name:   "empty game",
			gameID: uuid.New(),
			setupMock: func(mock sqlmock.Sqlmock, gameID uuid.UUID) {
				// Return zero rows for an “empty game”
				rows := sqlmock.NewRows([]string{
					"id", "parent_id", "title", "order_index", "wip_limit", "type",
				})
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(gameID).
					WillReturnRows(rows)
			},
			expected: []models.Column{},
		},
		// {
		// 	name:   "single column",
		// 	gameID: uuid.New(),
		// 	expected: []models.Column{
		// 		{
		// 			ID:         uuid.New(),
		// 			Title:      "To Do",
		// 			OrderIndex: 0,
		// 			WIPLimit:   nil,
		// 			Type:       "active",
		// 			SubColumns: nil,
		// 			ParentID:   nil,
		// 		},
		// 	},
		// },
		// {
		// 	name:   "multiple columns",
		// 	gameID: uuid.New(),
		// 	expected: []models.Column{
		// 		{
		// 			ID:         uuid.New(),
		// 			Title:      "To Do",
		// 			OrderIndex: 0,
		// 			WIPLimit:   nil,
		// 			Type:       "queue",
		// 			SubColumns: nil,
		// 			ParentID:   nil,
		// 		},
		// 		{
		// 			ID:         uuid.New(),
		// 			Title:      "In Progress",
		// 			OrderIndex: 1,
		// 			WIPLimit:   nil,
		// 			Type:       "active",
		// 			SubColumns: nil,
		// 			ParentID:   nil,
		// 		},
		// 		{
		// 			ID:         uuid.New(),
		// 			Title:      "Done",
		// 			OrderIndex: 2,
		// 			WIPLimit:   nil,
		// 			Type:       "done",
		// 			SubColumns: nil,
		// 			ParentID:   nil,
		// 		},
		// 	},
		// },
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			repo := columns.NewSQLRepo(db)
			tc.setupMock(mock, tc.gameID)

			columns, _ := repo.GetColumnsByGameID(context.Background(), tc.gameID)
			if tc.expected != nil {
				require.NotNil(t, columns)
			}
			require.Len(t, columns, len(tc.expected))

			for i, col := range columns {
				exp := tc.expected[i]
				require.Equal(t, exp.ID, col.ID, "ID mismatch at index %d", i)
				require.Equal(t, exp.Title, col.Title, "Title mismatch at index %d", i)
				require.Equal(t, exp.OrderIndex, col.OrderIndex, "OrderIndex mismatch at index %d", i)
				require.Equal(t, exp.Type, col.Type, "Type mismatch at index %d", i)
				require.Equal(t, exp.WIPLimit, col.WIPLimit, "WIPLimit mismatch at index %d", i)
				require.Equal(t, exp.ParentID, col.ParentID, "ParentID mismatch at index %d", i)
				// Check if the column matches the expected structure
				// This is a more detailed check, but you can adjust based on your needs
				// if col.ID != tc.expected[i].ID || col.Title != tc.expected[i].Title ||
				// 	col.OrderIndex != tc.expected[i].OrderIndex || col.Type != tc.expected[i].Type ||
				// 	col.WIPLimit != tc.expected[i].WIPLimit || col.ParentID != tc.expected[i].ParentID {
				// 	t.Errorf("column mismatch at index %d: got %+v, want %+v", i, col, tc.expected[i])
			}
			require.NoError(t, mock.ExpectationsWereMet(), "there are unmet expectations")
		})
	}
}
