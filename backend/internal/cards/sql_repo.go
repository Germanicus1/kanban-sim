package cards

import (
	"context"
	"database/sql"

	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/google/uuid"
)

type sqlRepo struct {
	db *sql.DB
}

func (r *sqlRepo) GetCardsByGameID(ctx context.Context, gameID uuid.UUID) ([]models.Card, error) {
	query := `SELECT id, game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day, order_index FROM cards WHERE game_id = $1`

	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card

	for rows.Next() {
		var c models.Card
		if err := rows.Scan(&c.ID, &c.GameID, &c.ColumnID, &c.Title, &c.ClassOfService, &c.ValueEstimate, &c.SelectedDay, &c.DeployedDay, &c.OrderIndex); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}
