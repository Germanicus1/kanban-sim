package cards_test

import (
	"context"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Germanicus1/kanban-sim/backend/internal/cards"
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

func TestGetCardsByGameID(t *testing.T) {
	const query = `SELECT id, game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day, order_index FROM cards WHERE game_id = $1`
	gameID := uuid.New()
	cardID := uuid.New()
	colID := uuid.New()
	ctx := context.Background()

	test := []struct {
		name            string
		setupMock       func(mock sqlmock.Sqlmock)
		wantErrContains string
		wantCards       *[]models.Card
	}{
		{
			name: "Success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WithArgs(gameID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "game_id", "column_id", "title", "class_of_service", "value_estimate", "selected_day", "deployed_day", "order_index"}).
						AddRow(cardID, gameID, colID, "Test Card", "Standard", 5, 1, 2, 0))
			},
			wantErrContains: "",
			wantCards: &[]models.Card{
				{
					ID:             cardID,
					GameID:         gameID,
					ColumnID:       colID,
					Title:          "Test Card",
					ClassOfService: "Standard",
					ValueEstimate:  5,
					SelectedDay:    1,
					DeployedDay:    2,
					OrderIndex:     0,
				},
			},
		},
	}
	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			defer db.Close()

			repo := cards.NewSQLRepo(db)
			tc.setupMock(mock)

			cards, err := repo.GetCardsByGameID(ctx, gameID)
			if err != nil {
				if tc.wantErrContains == "" || !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}

			if tc.wantCards == nil {
				t.Fatal("expected cards to be nil")
			}

			if !reflect.DeepEqual(cards, *tc.wantCards) {
				t.Errorf("expected cards %v, got %v", *tc.wantCards, cards)
			}

		})
	}
}
