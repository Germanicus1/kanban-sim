package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Germanicus1/kanban-sim/internal"
	"github.com/google/uuid"
)

type Card struct {
	ID             uuid.UUID `json:"id"`
	GameID         uuid.UUID `json:"game_id"`
	Title          string    `json:"title"`
	CardColumn     string    `json:"card_column"`
	ClassOfService string    `json:"class_of_service,omitempty"`
	ValueEstimate  string    `json:"value_estimate,omitempty"`
	EffortAnalysis int       `json:"effort_analysis,omitempty"`
	EffortDev      int       `json:"effort_development,omitempty"`
	EffortTest     int       `json:"effort_test,omitempty"`
	SelectedDay    int       `json:"selected_day,omitempty"`
	DeployedDay    int       `json:"deployed_day,omitempty"`
}

// CreateCard creates a new card
func CreateCard(w http.ResponseWriter, r *http.Request) {
	var card Card

	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST_BODY")
		return
	}

	if card.GameID == uuid.Nil {
		internal.RespondWithError(w, http.StatusBadRequest, "GAME_ID_REQUIRED")
		return
	}

	query := `
		INSERT INTO cards (game_id, title, card_column, class_of_service, value_estimate, effort_analysis, effort_development, effort_test, selected_day, deployed_day)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	err := internal.DB.QueryRow(
		query,
		card.GameID,
		card.Title,
		card.CardColumn,
		card.ClassOfService,
		card.ValueEstimate,
		card.EffortAnalysis,
		card.EffortDev,
		card.EffortTest,
		card.SelectedDay,
		card.DeployedDay,
	).Scan(&card.ID)

	if err != nil {
		status, errCode := internal.MapPostgresError(err)
		internal.RespondWithError(w, status, errCode)
		return
	}

	internal.RespondWithData(w, card)
}

// GetCard retrieves a card by ID
func GetCard(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	parsedID, err := uuid.Parse(id)
	if err != nil {
		internal.RespondWithError(w, http.StatusBadRequest, "INVALID_CARD_ID")
		return
	}

	log.Printf("Looking for card with ID: %s", parsedID)

	var card Card
	query := `SELECT id, game_id, title, card_column FROM cards WHERE id = $1`
	err = internal.DB.QueryRow(query, parsedID).Scan(&card.ID, &card.GameID, &card.Title, &card.CardColumn)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Card not found with ID: %s", parsedID)
			internal.RespondWithError(w, http.StatusNotFound, "CARD_NOT_FOUND")
		} else {
			log.Printf("Database error: %v", err)
			internal.RespondWithError(w, http.StatusInternalServerError, "DATABASE_ERROR")
		}
		return
	}

	internal.RespondWithData(w, card)
}

// UpdateCard updates card details
func UpdateCard(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	var card Card
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		UPDATE cards SET 
			title = $1, 
			card_column = $2, 
			class_of_service = $3, 
			value_estimate = $4, 
			effort_analysis = $5, 
			effort_development = $6, 
			effort_test = $7, 
			selected_day = $8, 
			deployed_day = $9
		WHERE id = $10
	`
	_, err := internal.DB.Exec(query, card.Title, card.CardColumn, card.ClassOfService, card.ValueEstimate, card.EffortAnalysis, card.EffortDev, card.EffortTest, card.SelectedDay, card.DeployedDay, id)
	if err != nil {
		http.Error(w, "Failed to update card", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteCard deletes a card by ID
func DeleteCard(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM cards WHERE id = $1`
	result, err := internal.DB.Exec(query, id)
	if err != nil {
		http.Error(w, "Failed to delete card", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
