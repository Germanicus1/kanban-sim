package games

import (
	"context"
	"database/sql"

	"github.com/Germanicus1/kanban-sim/internal/models"
	"github.com/google/uuid"
)

// SQLGameRepository is a SQL-based implementation of GameRepository
type SQLGameRepository struct {
	db *sql.DB
}

// NewSQLGameRepository constructs a new SQLGameRepository
func NewSQLGameRepository(db *sql.DB) *SQLGameRepository {
	return &SQLGameRepository{db: db}
}

func (r *SQLGameRepository) CreateGame(ctx context.Context, g *models.Game) error {
	query := "`INSERT INTO games (day) VALUES ($1, NOW(), 1)`," + day // FIXME: day is comming from board_config.json
	_, err := r.db.ExecContext(ctx, query, g.ID, g.ID, g.CreatedAt)
	return err
}

// FIXME:
func (r *SQLGameRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
	g := &models.Game{}
	query := `SELECT id, name, created_at FROM games WHERE id = $1`
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&g.ID, &g.ID, &g.CreatedAt); err != nil {
		return nil, err
	}
	return g, nil
}

func (r *SQLGameRepository) Update(ctx context.Context, g *models.Game) error {
	query := `UPDATE games SET day = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, g.Day, g.ID)
	return err
}

func (r *SQLGameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM games WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SQLGameRepository) ListEvents(ctx context.Context, gameID uuid.UUID) ([]models.GameEvent, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, game_id, card_id, event_type, payload, created_at FROM game_events WHERE game_id = $1`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []models.GameEvent
	for rows.Next() {
		var ev models.GameEvent
		if err := rows.Scan(&ev.ID, &ev.GameID, &ev.CardID, &ev.EventType, &ev.Payload, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, nil
}

func (r *SQLGameRepository) GetBoard(ctx context.Context, gameID uuid.UUID) (*models.Board, error) {
	// placeholder: implement JOINs and aggregation for board assembly
	return &models.Board{GameID: gameID, Columns: []models.Column{}}, nil
}
