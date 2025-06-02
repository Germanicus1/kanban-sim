package columns

import (
	"context"
	"database/sql"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type sqlRepo struct {
	db *sql.DB
}

func (r sqlRepo) GetColumnsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Column, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, wip_limit, col_type, parent_id, order_index
           FROM columns
          WHERE game_id = $1
       ORDER BY parent_id, order_index`,
		gameID, // pass UUID directly, not a string
	)
	cols := make([]models.Column, 0) // non‐nil from the start
	if err != nil {
		return cols, err
	}
	defer rows.Close()

	for rows.Next() {
		var c models.Column
		if err := rows.Scan(&c.ID, &c.Title, &c.WIPLimit, &c.Type, &c.ParentID, &c.OrderIndex); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}
