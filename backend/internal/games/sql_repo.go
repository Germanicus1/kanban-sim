package games

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

type sqlRepo struct {
	db *sql.DB
}

// CreateGame inserts a new game row, then seeds effort_types,
// columns (and subcolumns), cards and their efforts, all in one TX.
func (r *sqlRepo) CreateGame(ctx context.Context, cfg models.BoardConfig) (uuid.UUID, error) {
	// 1) begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	// 2) let Postgres create the game ID and return it
	var gameID uuid.UUID
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO games (created_at, day)
             VALUES (NOW(), 1)
         RETURNING id`,
	).Scan(&gameID); err != nil {
		tx.Rollback()
		return uuid.Nil, fmt.Errorf("insert game: %w", err)
	}

	// 3) seed effort_types, grabbing each new ID
	effortTypeIDs := make(map[string]uuid.UUID, len(cfg.EffortTypes))
	for idx, et := range cfg.EffortTypes {
		var etID uuid.UUID
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO effort_types (game_id, title, order_index)
                 VALUES ($1, $2, $3)
             RETURNING id`,
			gameID, et.Title, idx,
		).Scan(&etID); err != nil {
			tx.Rollback()
			return uuid.Nil, fmt.Errorf("insert effort type %q: %w", et.Title, err)
		}
		effortTypeIDs[et.Title] = etID
	}

	// 4) seed columns & subcolumns, grabbing each new ID
	columnIDs := make(map[string]uuid.UUID, len(cfg.Columns)*2)
	for _, col := range cfg.Columns {
		var mainID uuid.UUID
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO columns (game_id, title, parent_id, order_index)
                 VALUES ($1, $2, NULL, $3)
             RETURNING id`,
			gameID, col.Title, col.OrderIndex,
		).Scan(&mainID); err != nil {
			tx.Rollback()
			return uuid.Nil, fmt.Errorf("insert column %q: %w", col.Title, err)
		}
		columnIDs[col.Title] = mainID

		for _, sub := range col.SubColumns {
			var subID uuid.UUID
			if err := tx.QueryRowContext(ctx,
				`INSERT INTO columns (game_id, title, parent_id, order_index)
                     VALUES ($1, $2, $3, $4)
                 RETURNING id`,
				gameID, sub.Title, mainID, sub.OrderIndex,
			).Scan(&subID); err != nil {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("insert subcolumn %q under %q: %w",
					sub.Title, col.Title, err,
				)
			}
			columnIDs[col.Title+" - "+sub.Title] = subID
		}
	}

	// 5) seed cards & their efforts, grabbing each new ID
	for _, c := range cfg.Cards {
		colID, ok := columnIDs[c.ColumnTitle]
		if !ok {
			tx.Rollback()
			return uuid.Nil, fmt.Errorf("unknown column %q", c.ColumnTitle)
		}

		var cardID uuid.UUID
		if err := tx.QueryRowContext(ctx,
			`INSERT INTO cards
               (game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day)
             VALUES ($1,$2,$3,$4,$5,$6,$7)
         RETURNING id`,
			gameID, colID,
			c.Title, c.ClassOfService, c.ValueEstimate,
			c.SelectedDay, c.DeployedDay,
		).Scan(&cardID); err != nil {
			tx.Rollback()
			return uuid.Nil, fmt.Errorf("insert card %q: %w", c.Title, err)
		}

		for _, e := range c.Efforts {
			etID, ok := effortTypeIDs[e.EffortType]
			if !ok {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("unknown effort type %q", e.EffortType)
			}
			// you may not need the effortID, but we can capture it if you do
			var effortID uuid.UUID
			if err := tx.QueryRowContext(ctx,
				`INSERT INTO efforts (card_id, effort_type_id, estimate, remaining, actual)
                     VALUES ($1,$2,$3,$3,0)
                 RETURNING id`,
				cardID, etID, e.Estimate,
			).Scan(&effortID); err != nil {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("insert effort %q for card %q: %w",
					e.EffortType, c.Title, err,
				)
			}
		}
	}

	// 6) commit
	if err = tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("commit tx: %w", err)
	}
	return gameID, nil
}

// GetBoard loads an entire board for a given game ID.
func (r *sqlRepo) GetBoard(ctx context.Context, gameID uuid.UUID) (models.Board, error) {
	var board models.Board
	board.GameID = gameID

	// 1) Load all columns
	colRows, err := r.db.QueryContext(ctx, `
        SELECT id, parent_id, title, order_index
          FROM columns
         WHERE game_id = $1
         ORDER BY order_index
    `, gameID)
	if err != nil {
		return board, fmt.Errorf("query columns: %w", err)
	}
	defer colRows.Close()

	// Scan into flat slice
	cols := make([]models.Column, 0)
	for colRows.Next() {
		var c models.Column
		if err := colRows.Scan(
			&c.ID,
			&c.ParentID,
			&c.Title,
			&c.OrderIndex,
		); err != nil {
			return board, fmt.Errorf("scan column: %w", err)
		}
		cols = append(cols, c)
	}
	if err := colRows.Err(); err != nil {
		return board, fmt.Errorf("iterate columns: %w", err)
	}

	// 2) Build map of ID → *Column
	colMap := make(map[uuid.UUID]*models.Column, len(cols))
	for i := range cols {
		colMap[cols[i].ID] = &cols[i]
	}

	// 3) Attach subcolumns to parents
	for i := range cols {
		c := &cols[i]
		if c.ParentID != nil {
			parent := colMap[*c.ParentID]
			parent.SubColumns = append(parent.SubColumns, *c)
		}
	}

	// 4) Collect only top-level columns (ParentID == nil)
	for _, c := range cols {
		if c.ParentID == nil {
			board.Columns = append(board.Columns, *colMap[c.ID])
		}
	}

	// 5) Load effort types
	etRows, err := r.db.QueryContext(ctx, `
        SELECT id, title, order_index
          FROM effort_types
         WHERE game_id = $1
         ORDER BY order_index
    `, gameID)
	if err != nil {
		return board, fmt.Errorf("query effort_types: %w", err)
	}
	defer etRows.Close()

	for etRows.Next() {
		var et models.EffortType
		if err := etRows.Scan(&et.ID, &et.Title, &et.OrderIndex); err != nil {
			return board, fmt.Errorf("scan effort_type: %w", err)
		}
		board.EffortTypes = append(board.EffortTypes, et)
	}
	if err := etRows.Err(); err != nil {
		return board, fmt.Errorf("iterate effort_types: %w", err)
	}

	// 6) Load cards and their efforts
	cardRows, err := r.db.QueryContext(ctx, `
        SELECT id, game_id, column_id, title,
               class_of_service, value_estimate,
               selected_day, deployed_day
          FROM cards
         WHERE game_id = $1
         ORDER BY selected_day
    `, gameID)
	if err != nil {
		return board, fmt.Errorf("query cards: %w", err)
	}
	defer cardRows.Close()

	for cardRows.Next() {
		var c models.Card
		if err := cardRows.Scan(
			&c.ID,
			&c.GameID,
			&c.ColumnID,
			&c.Title,
			&c.ClassOfService,
			&c.ValueEstimate,
			&c.SelectedDay,
			&c.DeployedDay,
		); err != nil {
			return board, fmt.Errorf("scan card: %w", err)
		}

		// Load efforts for this card
		erRows, err := r.db.QueryContext(ctx, `
            SELECT et.title, e.estimate, e.remaining, e.actual
              FROM efforts e
              JOIN effort_types et ON et.id = e.effort_type_id
             WHERE e.card_id = $1
             ORDER BY et.order_index
        `, c.ID)
		if err != nil {
			return board, fmt.Errorf("query efforts: %w", err)
		}
		for erRows.Next() {
			var e models.Effort
			var rem, act sql.NullInt64
			if err := erRows.Scan(&e.EffortType, &e.Estimate, &rem, &act); err != nil {
				erRows.Close()
				return board, fmt.Errorf("scan effort: %w", err)
			}
			if rem.Valid {
				v := int(rem.Int64)
				e.Remaining = &v
			}
			if act.Valid {
				v := int(act.Int64)
				e.Actual = &v
			}
			c.Efforts = append(c.Efforts, e)
		}
		erRows.Close()

		board.Cards = append(board.Cards, c)
	}
	if err := cardRows.Err(); err != nil {
		return board, fmt.Errorf("iterate cards: %w", err)
	}

	return board, nil
}

func (r *sqlRepo) GetGameByID(ctx context.Context, id uuid.UUID) (models.Game, error) {
	const q = `SELECT id, created_at, day FROM games WHERE id = $1`
	var g models.Game

	switch err := r.db.QueryRowContext(ctx, q, id).Scan(&g.ID, &g.CreatedAt, &g.Day); err {
	case nil:
		return g, nil
	case sql.ErrNoRows:
		return models.Game{}, ErrNotFound
	default:
		return models.Game{}, err
	}
}

func (r *sqlRepo) DeleteGame(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM games WHERE id = $1`
	if _, err := r.db.ExecContext(ctx, q, id); err != nil {
		if err == sql.ErrNoRows {
			return response.ErrNotFound
		}
		return fmt.Errorf("delete game: %w", err)
	}
	return nil
}

func (r *sqlRepo) UpdateGame(ctx context.Context, id uuid.UUID, day int) error {
	const q = `UPDATE games SET day = $1 WHERE id = $2`
	if _, err := r.db.ExecContext(ctx, q, day, id); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("update game: %w", err)
	}
	return nil
}
