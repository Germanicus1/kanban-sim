package games

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Germanicus1/kanban-sim/internal/models"
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

	// 2) insert game
	gameID := uuid.New()
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO games (id, created_at, day) VALUES ($1, NOW(), 1)`,
		gameID,
	); err != nil {
		tx.Rollback()
		return uuid.Nil, fmt.Errorf("insert game: %w", err)
	}

	// 3) seed effort_types
	effortTypeIDs := make(map[string]uuid.UUID, len(cfg.EffortTypes))
	for idx, et := range cfg.EffortTypes {
		etID := uuid.New()
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO effort_types (id, game_id, title, order_index)
             VALUES ($1,$2,$3,$4)`,
			etID, gameID, et.Title, idx,
		); err != nil {
			tx.Rollback()
			return uuid.Nil, fmt.Errorf("insert effort type %q: %w", et.Title, err)
		}
		effortTypeIDs[et.Title] = etID
	}

	// 4) seed columns & subcolumns
	columnIDs := make(map[string]uuid.UUID, len(cfg.Columns)*2)
	for _, col := range cfg.Columns {
		mainID := uuid.New()
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO columns (id, game_id, title, parent_id, order_index)
             VALUES ($1,$2,$3,NULL,$4)`,
			mainID, gameID, col.Title, col.OrderIndex,
		); err != nil {
			tx.Rollback()
			return uuid.Nil, fmt.Errorf("insert column %q: %w", col.Title, err)
		}
		columnIDs[col.Title] = mainID

		for _, sub := range col.SubColumns {
			subID := uuid.New()
			if _, err = tx.ExecContext(ctx,
				`INSERT INTO columns (id, game_id, title, parent_id, order_index)
                 VALUES ($1,$2,$3,$4,$5)`,
				subID, gameID, sub.Title, mainID, sub.OrderIndex,
			); err != nil {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("insert subcolumn %q under %q: %w",
					sub.Title, col.Title, err,
				)
			}
			columnIDs[col.Title+" - "+sub.Title] = subID
		}
	}

	// 5) seed cards & their efforts
	for _, c := range cfg.Cards {
		cardID := uuid.New()
		colID, ok := columnIDs[c.ColumnTitle]
		if !ok {
			tx.Rollback()
			return uuid.Nil, fmt.Errorf("unknown column %q", c.ColumnTitle)
		}

		if _, err = tx.ExecContext(ctx,
			`INSERT INTO cards
               (id, game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day)
             VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
			cardID, gameID, colID,
			c.Title, c.ClassOfService, c.ValueEstimate,
			c.SelectedDay, c.DeployedDay,
		); err != nil {
			tx.Rollback()
			return uuid.Nil, fmt.Errorf("insert card %q: %w", c.Title, err)
		}

		for _, e := range c.Efforts {
			etID, ok := effortTypeIDs[e.EffortType]
			if !ok {
				tx.Rollback()
				return uuid.Nil, fmt.Errorf("unknown effort type %q", e.EffortType)
			}
			if _, err = tx.ExecContext(ctx,
				`INSERT INTO efforts (id, card_id, effort_type_id, estimate, remaining, actual)
                 VALUES($1,$2,$3,$4,$4,0)`,
				uuid.New(), cardID, etID, e.Estimate,
			); err != nil {
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
