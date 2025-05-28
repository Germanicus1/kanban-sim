package players_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/Germanicus1/kanban-sim/internal/players"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSQLRepo_CreatePlayer_Success(t *testing.T) {
	// 1) spin up a sqlmock database + mock controller
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	// 2) make our repo against the fake DB
	repo := players.NewSQLRepo(db)

	// 3) set up the expected calls in order:
	//
	//    a) BeginTx
	mock.ExpectBegin()

	//    b) the INSERT; note: we match the SQL via regexp,
	//       and assert the two args (name & game_id)
	newGameID := uuid.New()
	playerName := "Alice"
	mock.
		ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO players (name, game_id, created_at)
             VALUES ($1, $2, NOW())
             RETURNING id`,
		)).
		WithArgs(playerName, newGameID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).
			AddRow(uuid.New()),
		)

	//    c) Commit
	mock.ExpectCommit()

	// 4) call the method under test
	outID, err := repo.CreatePlayer(
		context.Background(),
		models.Player{
			Name:   playerName,
			GameID: newGameID,
		},
	)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, outID, "expected a non-nil UUID")

	// 5) assert all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())
}

// TODO: Implement the rest of the methods for sqlRepo
func TestRepository_GetPlayerByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := players.NewSQLRepo(db)
	mock.ExpectBegin()
	newPlayerID := uuid.New()
	gameID := uuid.New()
	playerName := "Alice"

	mock.
		ExpectQuery(regexp.QuoteMeta(
			`SELECT id, name, game_id FROM players WHERE id = $1`,
		)).
		WithArgs(newPlayerID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "game_id"}).
			AddRow(newPlayerID, playerName, gameID),
		)
	mock.ExpectCommit()
	player, err := repo.GetPlayerByID(context.Background(), newPlayerID)
	require.NoError(t, err)
	require.Equal(t, newPlayerID, player.ID)
	require.Equal(t, playerName, player.Name)
	require.Equal(t, gameID, player.GameID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Update(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := players.NewSQLRepo(db)
	playerID := uuid.New()
	playerName := "Alice"

	mock.
		ExpectExec(regexp.QuoteMeta(
			`UPDATE players SET name = $1 WHERE id = $2`,
		)).
		WithArgs(playerName, playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdatePlayer(context.Background(), playerID, playerName)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepository_Delete(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := players.NewSQLRepo(db)
	playerID := uuid.New()
	mock.
		ExpectExec(regexp.QuoteMeta(
			`DELETE FROM players WHERE id = $1`,
		)).
		WithArgs(playerID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	err := repo.DeletePlayer(context.Background(), playerID)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
