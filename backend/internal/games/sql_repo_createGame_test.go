package games

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

func TestSQLRepo_CreateGame(t *testing.T) {
	gameID := uuid.New()
	cardID := uuid.New()
	colID := uuid.New()
	subColID := uuid.New()
	etID := uuid.New()
	ctx := context.Background()

	type args struct {
		ctx context.Context
		cfg models.BoardConfig
	}
	tests := []struct {
		name    string
		args    args
		want    uuid.UUID
		wantErr bool
		prepare func(sqlmock.Sqlmock, models.BoardConfig)
	}{
		{
			name:    "success minimal config",
			args:    args{ctx: ctx, cfg: models.BoardConfig{}},
			want:    gameID,
			wantErr: false,
			prepare: func(m sqlmock.Sqlmock, cfg models.BoardConfig) {
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO games .* RETURNING id`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(gameID))
				m.ExpectCommit()
			},
		},
		{
			name: "success full config",
			args: args{
				ctx: ctx,
				cfg: models.BoardConfig{
					EffortTypes: []models.EffortType{
						{Title: "development", OrderIndex: 1},
					},
					Columns: []models.Column{
						{
							ParentID:   nil,
							Title:      "development",
							OrderIndex: 0,
							WIPLimit:   3,
							Type:       "active",
							SubColumns: []models.Column{
								{Title: "in progress", OrderIndex: 0},
								{Title: "ready", OrderIndex: 1},
							},
						},
					},
					Cards: []models.Card{
						{
							// ColumnTitle must match "development - ready"
							Title:          "some-title", // Title is unused by SQLRepo; ColumnTitle is used instead
							ColumnTitle:    "development - ready",
							ClassOfService: "standard",
							ValueEstimate:  "high",
							SelectedDay:    1,
							DeployedDay:    2,
							OrderIndex:     0,
							Efforts: []models.Effort{
								{EffortType: "development", Estimate: 3},
							},
						},
					},
				},
			},
			want:    gameID,
			wantErr: false,
			prepare: func(m sqlmock.Sqlmock, cfg models.BoardConfig) {
				developmentID := colID // main column “development”
				inProgID := subColID   // subcolumn “in progress”
				readyID := subColID    // subcolumn “ready”

				// 1) BEGIN
				m.ExpectBegin()

				// 2) INSERT INTO games … RETURNING id
				m.ExpectQuery(`INSERT INTO games .* RETURNING id`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(gameID))

				// 3) INSERT INTO effort_types … RETURNING id
				m.ExpectQuery(`INSERT INTO effort_types .* RETURNING id`).
					WithArgs(gameID, "development", 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(etID))

				// 4) INSERT main column “development”
				m.ExpectQuery(`INSERT INTO columns .* RETURNING id`).
					WithArgs(
						gameID,        // $1 → game_id
						"development", // $2 → title
						0,             // $3 → order_index
						3,             // $4 → wip_limit
						"active",      // $5 → col_type
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(developmentID))

				// 5) INSERT subcolumn “in progress” under “development”
				m.ExpectQuery(`INSERT INTO columns .* RETURNING id`).
					WithArgs(
						gameID,        // $1 → game_id
						"in progress", // $2 → title
						developmentID, // $3 → parent_id
						0,             // $4 → order_index
						0,             // $5 → wip_limit (default)
						"queue",       // $6 → col_type (default)
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(inProgID))

				// 6) INSERT subcolumn “ready” under “development”
				m.ExpectQuery(`INSERT INTO columns .* RETURNING id`).
					WithArgs(
						gameID,        // $1 → game_id
						"ready",       // $2 → title
						developmentID, // $3 → parent_id
						1,             // $4 → order_index
						0,             // $5 → wip_limit (default)
						"queue",       // $6 → col_type (default)
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(readyID))

				// 7) INSERT INTO cards (game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day)
				m.ExpectQuery(`INSERT INTO cards .* RETURNING id`).
					WithArgs(
						gameID,       // $1 → game_id
						readyID,      // $2 → column_id
						"some-title", // $3 → title
						"standard",   // $4 → class_of_service
						"high",       // $5 → value_estimate
						1,            // $6 → selected_day
						2,            // $7 → deployed_day
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(cardID))

				// 8) INSERT INTO efforts (card_id, effort_type_id, estimate, remaining, actual)
				m.ExpectQuery(`INSERT INTO efforts .* RETURNING id`).
					WithArgs(cardID, etID, 3).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

				// 9) COMMIT
				m.ExpectCommit()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: open sqlmock DB
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating sqlmock: %v", err)
			}
			defer db.Close()

			// Set up the expected queries & rows
			tc.prepare(mock, tc.args.cfg)

			// Act: call CreateGame on our repo backed by sqlmock
			r := NewSQLRepo(db)
			got, err := r.CreateGame(tc.args.ctx, tc.args.cfg)

			// Assert: error expectation
			if (err != nil) != tc.wantErr {
				t.Errorf("CreateGame() error = %v, wantErr %v", err, tc.wantErr)
			}
			// Assert: returned gameID
			if got != tc.want {
				t.Errorf("CreateGame() = %v, want %v", got, tc.want)
			}
			// Verify all expected SQL calls were made
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}
