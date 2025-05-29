package games

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

func TestSQLRepo_CreateGame(t *testing.T) {
	testGameID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testEffTypeID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testColumnID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testSubcolumnID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testCardID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	testEffortID := uuid.MustParse("66666666-6666-6666-6666-666666666666")

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
			args:    args{ctx: context.Background(), cfg: models.BoardConfig{}},
			want:    testGameID,
			wantErr: false,
			prepare: func(m sqlmock.Sqlmock, cfg models.BoardConfig) {
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO games .* RETURNING id`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testGameID))
				m.ExpectCommit()
			},
		},

		{
			name: "success full config",
			args: args{
				ctx: context.Background(),
				cfg: models.BoardConfig{
					EffortTypes: []models.EffortType{{Title: "Bug", OrderIndex: 0}},
					Columns: []models.Column{
						{Title: "Backlog", OrderIndex: 0, SubColumns: []models.Column{{Title: "Ready", OrderIndex: 0}}},
					},
					Cards: []models.Card{
						{
							// composite key matching code: <column.Title> + " - " + <sub.Title>
							ColumnTitle:    "Backlog - Ready",
							Title:          "Implement feature",
							ClassOfService: "Standard",
							ValueEstimate:  "5",
							SelectedDay:    1,
							DeployedDay:    0,
							Efforts:        []models.Effort{{EffortType: "Bug", Estimate: 3}},
						},
					},
				},
			},
			want:    testGameID,
			wantErr: false,
			prepare: func(m sqlmock.Sqlmock, cfg models.BoardConfig) {
				m.ExpectBegin()

				// Insert game
				m.ExpectQuery(`INSERT INTO games .* RETURNING id`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testGameID))

				// Insert effort type
				m.ExpectQuery(`INSERT INTO effort_types .* RETURNING id`).
					WithArgs(testGameID, "Bug", 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testEffTypeID))

				// Insert main column
				m.ExpectQuery(`INSERT INTO columns .* RETURNING id`).
					WithArgs(testGameID, "Backlog", 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testColumnID))

				// Insert subcolumn
				m.ExpectQuery(`INSERT INTO columns .* RETURNING id`).
					WithArgs(testGameID, "Ready", testColumnID, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testSubcolumnID))

				// Insert card into subcolumn
				m.ExpectQuery(`INSERT INTO cards .* RETURNING id`).
					WithArgs(testGameID, testSubcolumnID, "Implement feature", "Standard", "5", 1, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testCardID))

				// Insert effort
				m.ExpectQuery(`INSERT INTO efforts .* RETURNING id`).
					WithArgs(testCardID, testEffTypeID, 3).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testEffortID))

				m.ExpectCommit()
			},
		},

		{
			name: "error unknown column",
			args: args{ctx: context.Background(), cfg: models.BoardConfig{
				Cards: []models.Card{{ColumnTitle: "Nonexistent", Title: "Test", ClassOfService: "Standard", ValueEstimate: "1", SelectedDay: 1, DeployedDay: 0, Efforts: nil}},
			}},
			want:    uuid.Nil,
			wantErr: true,
			prepare: func(m sqlmock.Sqlmock, cfg models.BoardConfig) {
				m.ExpectBegin()
				m.ExpectQuery(`INSERT INTO games .* RETURNING id`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(testGameID))
				m.ExpectRollback()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating sqlmock: %v", err)
			}
			defer db.Close()
			// set up expectations
			tt.prepare(mock, tt.args.cfg)

			// instantiate repository via public constructor
			r := NewSQLRepo(db)
			got, err := r.CreateGame(tt.args.ctx, tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGame() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("CreateGame() = %v, want %v", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}
