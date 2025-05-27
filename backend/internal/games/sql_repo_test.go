// internal/games/sql_repo_test.go
package games

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSQLRepo_GetGameByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSQLRepo(db)
	id := uuid.New()
	createdAt := "2025-05-15T00:00:00Z"
	day := 1

	// Expect the query and return one row
	rows := sqlmock.NewRows([]string{"id", "created_at", "day"}).
		AddRow(id, createdAt, day)
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, created_at, day FROM games WHERE id = $1"),
	).
		WithArgs(id).
		WillReturnRows(rows)

	g, err := repo.GetGameByID(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, id, g.ID)
	require.Equal(t, createdAt, g.CreatedAt)
	require.Equal(t, day, g.Day)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLRepo_GetGameByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSQLRepo(db)
	id := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, created_at, day FROM games WHERE id = $1"),
	).
		WithArgs(id).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.GetGameByID(context.Background(), id)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestSQLRepo_GetBoard_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewSQLRepo(db)
	id := uuid.New()

	// 1) columns returns no rows
	mock.ExpectQuery("SELECT id, parent_id, title, order_index FROM columns").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "parent_id", "title", "order_index"}))

	// 2) effort_types returns no rows
	mock.ExpectQuery("SELECT id, title, order_index FROM effort_types").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "order_index"}))

	// 3) cards returns no rows
	mock.ExpectQuery("SELECT id, game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day FROM cards").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "game_id", "column_id", "title", "class_of_service", "value_estimate", "selected_day", "deployed_day"}))

	board, err := repo.GetBoard(context.Background(), id)
	require.NoError(t, err)
	// empty board: only GameID set
	require.Equal(t, id, board.GameID)
	require.Empty(t, board.Columns)
	require.Empty(t, board.EffortTypes)
	require.Empty(t, board.Cards)

	require.NoError(t, mock.ExpectationsWereMet())
}

// internal/games/sql_repo_test.go
func TestSQLRepo_DeleteGame_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewSQLRepo(db)
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM games WHERE id = $1"),
	).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteGame(context.Background(), id)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLRepo_DeleteGame_NotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewSQLRepo(db)
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(
		"DELETE FROM games WHERE id = $1"),
	).WithArgs(id).WillReturnError(sql.ErrNoRows)

	err := repo.DeleteGame(context.Background(), id)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestSQLRepo_UpdateGame_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := NewSQLRepo(db)
	id := uuid.New()

	mock.ExpectExec(regexp.QuoteMeta(
		"UPDATE games SET day = $1 WHERE id = $2"),
	).WithArgs(1, id).WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.UpdateGame(context.Background(), id, 1)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
