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

		//FIXME: this test is not complete, it needs to insert the card and its efforts
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
							// NOTICE: c.Title must match the key in columnIDs,
							// which is formed as "<col.Title> - <sub.Title>".
							// Because we want this card to go under the "Ready" subcolumn
							// of the column "development", we set:
							Title:          "development - ready",
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
				// We will produce deterministic IDs for each INSERT:
				//   - colID    = ID of main column “development”
				//   - inProgID = ID of subcolumn “in progress”
				//   - readyID  = ID of subcolumn “Ready”
				developmentID := colID // alias for main “development”
				inProgID := subColID   // will be returned for “in progress”
				readyID := subColID    // will be returned for “ready”

				// 1) BEGIN
				m.ExpectBegin()

				// 2) INSERT INTO games (... RETURNING id)
				m.ExpectQuery(`INSERT INTO games .* RETURNING id`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(gameID))

				// 3) INSERT INTO effort_types (game_id, title, order_index) RETURNING id
				m.ExpectQuery(`INSERT INTO effort_types .* RETURNING id`).
					WithArgs(gameID, "development", 0).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(etID))

				// 4) INSERT main column “development”
				//    SQL in CreateGame:
				//      INSERT INTO columns
				//        (game_id, title, parent_id, order_index, wip_limit, col_type)
				//      VALUES ($1,$2,NULL,$3,$4,$5)
				//      RETURNING id
				m.ExpectQuery(`INSERT INTO columns .* RETURNING id`).
					WithArgs(
						gameID,        // $1 → game_id
						"development", // $2 → title
						0,             // $3 → order_index
						3,             // $4 → wip_limit (cfg.WIPLimit was 3)
						"active",      // $5 → col_type  (cfg.Type was "active")
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(developmentID))

				// 5) INSERT subcolumn “in progress” under “development”
				//    SQL:
				//      INSERT INTO columns
				//        (game_id, title, parent_id, order_index, wip_limit, col_type)
				//      VALUES ($1,$2,$3,$4,$5,$6)
				m.ExpectQuery(`INSERT INTO columns .* RETURNING id`).
					WithArgs(
						gameID,        // $1 → game_id
						"in progress", // $2 → title
						developmentID, // $3 → parent_id
						0,             // $4 → order_index
						0,             // $5 → wip_limit (no WIPLimit on sub → default 0)
						"queue",       // $6 → col_type (no Type on sub → default "queue")
					).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(inProgID))

				// 6) INSERT subcolumn “Ready” under “development”
				//   SQL:
				//     INSERT INTO columns
				//       (game_id, title, parent_id, order_index, wip_limit, col_type)
				//     VALUES ($1,$2,$3,$4,$5,$6)
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

				// 7) INSERT card into development - ready subcolumn SQL:
				//   INSERT INTO cards
				//      (game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day)
				//   VALUES ($1,$2,$3,$4,$5,$6,$7)
				m.ExpectQuery(`INSERT INTO cards .* RETURNING id`).
					WithArgs(
						gameID,                // $1 → game_id
						readyID,               // $2 → column_id (lookup from columnIDs["development - Ready"])
						"development - ready", // $3 → title  (equal to c.Title)
						"standard",            // $4 → class_of_service
						"high",                // $5 → value_estimate
						1,                     // $6 → selected_day
						2,                     // $7 → deployed_day
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
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("error creating sqlmock: %v", err)
			}
			defer db.Close()
			// set up expectations
			tc.prepare(mock, tc.args.cfg)

			// instantiate repository via public constructor
			r := NewSQLRepo(db)
			got, err := r.CreateGame(tc.args.ctx, tc.args.cfg)
			if (err != nil) != tc.wantErr {
				t.Errorf("CreateGame() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("CreateGame() = %v, want %v", got, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unmet expectations: %v", err)
			}
		})
	}
}
