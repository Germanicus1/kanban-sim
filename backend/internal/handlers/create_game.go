package handlers

import (
	"fmt"
	"net/http"

	"github.com/Germanicus1/kanban-sim/backend/internal/config"
	"github.com/Germanicus1/kanban-sim/backend/internal/database"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

// CreateGame creates a new game and seeds effort_types, columns, cards, and
// efforts.
func CreateGame(w http.ResponseWriter, r *http.Request) {
	// 0) load config
	cfg, err := config.LoadBoardConfig()
	_ = cfg
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"configuration load failed: "+err.Error())
		return
	}

	// 1) begin transaction
	tx, err := database.DB.Begin()
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"failed to begin transaction")
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after rollback
		} else if err != nil {
			tx.Rollback() // rollback only if there was an error
		}
	}()

	// 2) insert game
	gameID := uuid.New()
	if _, err := tx.Exec(
		`INSERT INTO games (id, created_at, day) VALUES ($1, NOW(), 1)`,
		gameID,
	); err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"failed to insert game: "+err.Error())
		return
	}

	// 3) seed effort_types
	effortTypeIDs := make(map[string]uuid.UUID, len(cfg.EffortTypes))
	for idx, et := range cfg.EffortTypes {
		etID := uuid.New()
		if _, err := tx.Exec(
			`INSERT INTO effort_types (id, game_id, title, order_index)
			VALUES ($1,$2,$3,$4)`,
			etID, gameID, et.Title, idx,
		); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError,
				"failed to insert effort type: "+err.Error())
			return
		}
		effortTypeIDs[et.Title] = etID
	}

	// 4) seed columns & subcolumns
	columnIDs := make(map[string]uuid.UUID, len(cfg.Columns)*2)
	for _, col := range cfg.Columns {
		mainID := uuid.New()
		if _, err := tx.Exec(
			`INSERT INTO columns (id, game_id, title, parent_id, order_index)
					VALUES ($1,$2,$3,NULL,$4)`,
			mainID, gameID, col.Title, col.OrderIndex,
		); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "failed to insert column: "+err.Error())
			return
		}
		columnIDs[col.Title] = mainID

		for _, sub := range col.SubColumns {
			subID := uuid.New()
			if _, err := tx.Exec(
				`INSERT INTO columns (id, game_id, title, parent_id, order_index)
				VALUES ($1,$2,$3,$4,$5)`,
				subID, gameID, sub.Title, mainID, sub.OrderIndex,
			); err != nil {
				msg := fmt.Sprintf("failed to insert subcolumn %q under %q: %v", sub.Title, col.Title, err)
				response.RespondWithError(w, http.StatusInternalServerError, msg)
				return
			}
			columnIDs[col.Title+" - "+sub.Title] = subID
		}
	}

	// REF  Seeding effort_types, columns, cards, and efforts

	// 5) seed cards & their efforts

	for _, c := range cfg.Cards {
		cardID := uuid.New()
		colID, ok := columnIDs[c.ColumnTitle]
		if !ok {
			response.RespondWithError(w, http.StatusInternalServerError, "unknown column "+colID.String())
			return
		}
		// _ = cardID
		// _ = colID
		if _, err := tx.Exec(
			`INSERT INTO cards
	   					(id, game_id, column_id, title, class_of_service, value_estimate, selected_day, deployed_day)
	   					VALUES($1,$2,$3,$4,$5,$6,$7,$8)`,
			cardID, gameID, colID,
			c.Title, c.ClassOfService, c.ValueEstimate,
			c.SelectedDay, c.DeployedDay,
		); err != nil {
			response.RespondWithError(w, http.StatusInternalServerError, "failed to insert card: "+err.Error())
			return
		}

		for _, e := range c.Efforts {
			etID, ok := effortTypeIDs[e.EffortType]
			if !ok {
				response.RespondWithError(w, http.StatusInternalServerError, "unknown effort type "+e.EffortType)
				return
			}
			if _, err := tx.Exec(
				`INSERT INTO efforts (id, card_id, effort_type_id, estimate, remaining, actual)
							VALUES($1,$2,$3,$4,$4,0)`,
				uuid.New(), cardID, etID, e.Estimate,
			); err != nil {
				response.RespondWithError(w, http.StatusInternalServerError, "failed to insert effort: "+err.Error())
				return
			}
		}
	}

	// 6) commit
	if err := tx.Commit(); err != nil {
		response.RespondWithError(w, http.StatusInternalServerError, "failed to commit transaction: "+err.Error())
		return
	}

	// 7) respond
	response.RespondWithData(w, map[string]uuid.UUID{"id": gameID})
}
