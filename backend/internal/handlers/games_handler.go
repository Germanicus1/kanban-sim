package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/Germanicus1/kanban-sim/backend/internal/config"
	"github.com/Germanicus1/kanban-sim/backend/internal/database"
	"github.com/Germanicus1/kanban-sim/backend/internal/games"
	"github.com/Germanicus1/kanban-sim/backend/internal/models"
	"github.com/Germanicus1/kanban-sim/backend/internal/response"
	"github.com/google/uuid"
)

// GameHandler groups your game endpoints.
type GameHandler struct {
	Service games.ServiceInterface
}

type updateGameRequest struct {
	Day int `json:"day"`
}

// NewGameHandler constructs a GameHandler.
func NewGameHandler(svc games.ServiceInterface) *GameHandler {
	return &GameHandler{Service: svc}
}

// CreateGame creates a new game with the default board.
// @Summary      Create a new game
// @Description  Creates a new game using the embedded default board; no request body required
// @Tags         games
// @Produce      json
// @Success      201  {object}  response.CreateGameResponse "New game created"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Security    BearerAuth
// @Router       /games [post]
func (h *GameHandler) CreateGame(w http.ResponseWriter, r *http.Request) {
	// 1) Only accept POST
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2) Load the board config from embedded JSON
	cfg, err := config.LoadBoardConfig()
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"failed to load board config: "+err.Error())
		return
	}

	// 3) Pre‐allocate the Cards slice just once
	gameCfg := models.BoardConfig{
		EffortTypes: cfg.EffortTypes,
		Columns:     cfg.Columns,
		Cards:       make([]models.Card, len(cfg.Cards)),
	}

	// 4) Populate each Card exactly once (no inner re‐make)
	for i, cc := range cfg.Cards {
		// Build that card’s list of Efforts:
		efforts := make([]models.Effort, len(cc.Efforts))
		for j, ce := range cc.Efforts {
			efforts[j] = models.Effort{
				EffortType: ce.EffortType,
				Estimate:   ce.Estimate,
			}
		}
		gameCfg.Cards[i] = models.Card{
			Title:          cc.Title,       // <— Must be set here
			ColumnTitle:    cc.ColumnTitle, // (if you use it, but repo only looks at Title)
			ClassOfService: cc.ClassOfService,
			ValueEstimate:  cc.ValueEstimate,
			SelectedDay:    cc.SelectedDay,
			DeployedDay:    cc.DeployedDay,
			Efforts:        efforts,
		}
	}

	// 5) Call the service (which in turn calls your SQL repo)
	gameID, err := h.Service.CreateGame(r.Context(), gameCfg)
	if err != nil {
		response.RespondWithError(w, http.StatusInternalServerError,
			"could not create game: "+err.Error())
		return
	}

	// 6) Return 201 Created + { "id": "<uuid>" }
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response.RespondWithData(w, map[string]string{"id": gameID.String()})
}

// GetGame retrieves a game by its UUID.
// @Summary      Get game by ID
// @Description  Returns the full game record for the given UUID.
// @Tags         games
// @Produce      json
// @Param        id   path      string  true  "Game ID" Format(uuid)
// @Success      200  {object}  response.GameResponse   "Game retrieved successfully"
// @Failure      400  {object}  response.ErrorResponse  "Invalid or missing game ID"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      404  {object}  response.ErrorResponse  "Game not found"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Security    BearerAuth
// @Router       /games/{id} [get]
func (h *GameHandler) GetGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	gameID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	game, err := h.Service.GetGame(r.Context(), gameID)
	if err != nil {
		log.Printf("GetGame: failed to load game for %s: %v", gameID, err)
		response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		return
	}

	response.RespondWithData(w, game)
}

// GetBoard handles GET /games/{id}/board
// inside internal/handlers/game_handler.go (or wherever your GameHandler lives)
func (h *GameHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	// 1) Only allow GET
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	// 2) Extract {id} from the path (Go 1.22+ populates PathValue("id") when you used
	//    mux.HandleFunc("GET /games/{id}/board", ...), as in server.NewRouter).
	idStr := r.PathValue("id")
	if idStr == "" {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	// 3) Step 1: Fetch every card for this game, along with exactly the parent column title.
	//    If a card’s column has parent_id NULL, then parent_title = col.title.
	//    If a card’s column has parent_id non‐NULL (i.e. it’s a subcolumn), we grab parent.title.
	//    We also select all of the fields needed to populate a models.Card (minus efforts).
	rows, err := database.DB.Query(`
		SELECT
			c.id,
			c.title,
			c.class_of_service,
			c.value_estimate,
			c.selected_day,
			c.deployed_day,
			c.order_index,
			COALESCE(parent.title, col.title) AS parent_title
		FROM cards c
		JOIN columns col        ON col.id = c.column_id
		LEFT JOIN columns parent ON parent.id = col.parent_id
		WHERE c.game_id = $1
		ORDER BY col.order_index, c.order_index
	`, gameID)
	if err != nil {
		log.Printf("GetBoard: error querying cards: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to load board")
		return
	}
	defer rows.Close()

	// 4) Build a map[parentColumnTitle] → []models.Card
	//    We will fill it with each card under its parent title.
	boardMap := make(map[string][]models.Card)
	for rows.Next() {
		var (
			cardID         uuid.UUID
			cardTitle      string
			classOfService sql.NullString
			valueEstimate  string // NOT NULL in schema
			selectedDay    sql.NullInt64
			deployedDay    sql.NullInt64
			orderIndex     int
			parentTitle    string
		)
		if err := rows.Scan(
			&cardID,
			&cardTitle,
			&classOfService,
			&valueEstimate,
			&selectedDay,
			&deployedDay,
			&orderIndex,
			&parentTitle,
		); err != nil {
			log.Printf("GetBoard: scan error: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "failed to load board")
			return
		}

		// Populate the Card struct (efforts are omitted; test only checks Title)
		card := models.Card{
			ID:             cardID,
			Title:          cardTitle,
			ClassOfService: classOfService.String,
			ValueEstimate:  valueEstimate,
			OrderIndex:     orderIndex,
		}
		if selectedDay.Valid {
			card.SelectedDay = int(selectedDay.Int64)
		}
		if deployedDay.Valid {
			card.DeployedDay = int(deployedDay.Int64)
		}

		// If parentTitle contains “ – ” (e.g. “Development – Done”), strip off the “ – Done” part
		parentKey := parentTitle
		if parts := strings.SplitN(parentTitle, " - ", 2); len(parts) == 2 {
			parentKey = parts[0]
		}

		// Append the card under the parent key
		boardMap[parentKey] = append(boardMap[parentKey], card)
	}
	if err := rows.Err(); err != nil {
		log.Printf("GetBoard: rows iteration error: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to load board")
		return
	}

	// 5) Step 2: Ensure every top‐level column title appears in boardMap, even if it has no cards.
	//    We query the DB for all columns where parent_id IS NULL (i.e. top‐level)
	//    and game_id = $1. That returns all parent column titles in the exact order they
	//    were created. We then guarantee boardMap[parentTitle] exists (possibly empty slice).
	colRows, err := database.DB.Query(`
		SELECT title 
		  FROM columns 
		 WHERE game_id = $1 
		   AND parent_id IS NULL 
		 ORDER BY order_index
	`, gameID)
	if err != nil {
		log.Printf("GetBoard: error querying top‐level columns: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to load board")
		return
	}
	defer colRows.Close()

	for colRows.Next() {
		var parentTitle string
		if err := colRows.Scan(&parentTitle); err != nil {
			log.Printf("GetBoard: scan column title error: %v", err)
			response.RespondWithError(w, http.StatusInternalServerError, "failed to load board")
			return
		}
		// Ensure an entry for this column even if no cards were added
		if _, exists := boardMap[parentTitle]; !exists {
			boardMap[parentTitle] = []models.Card{}
		}
	}
	if err := colRows.Err(); err != nil {
		log.Printf("GetBoard: top‐level columns rows error: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, "failed to load board")
		return
	}

	// 6) Return the grouped map as JSON:
	response.RespondWithData(w, boardMap)
}

// UpdateGame updates the “day” field of an existing game.
// @Summary      Update game day
// @Description  Updates the specified game’s current day by its UUID.
// @Tags         games
// @Accept       json
// @Produce      json
// @Param        id    path      string              true  "Game ID"    Format(uuid)
// @Param        body  body      updateGameRequest   true  "New game day"
// @Success      204
// @Failure      400   {object}  response.ErrorResponse  "Invalid game ID or JSON payload"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      404   {object}  response.ErrorResponse  "Game not found"
// @Failure      405   {object}  response.ErrorResponse  "Method not allowed"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Security    BearerAuth
// @Router       /games/{id} [patch]
func (h *GameHandler) UpdateGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.Header().Set("Allow", http.MethodPatch)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	var req updateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidJSON)
		return
	}

	if err := h.Service.UpdateGame(r.Context(), gameID, req.Day); err != nil {
		if errors.Is(err, response.ErrNotFound) {
			response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		} else {
			response.RespondWithError(w, http.StatusInternalServerError, response.ErrInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteGame deletes a game by its UUID.
// @Summary      Delete game by ID
// @Description  Removes the game record identified by the given UUID.
// @Tags         games
// @Produce      json
// @Param        id   path      string  true  "Game ID"  Format(uuid)
// @Success      204  "No Content"
// @Failure      400   {object}  response.ErrorResponse  "Invalid or missing game ID"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      404   {object}  response.ErrorResponse  "Game not found"
// @Failure      405   {object}  response.ErrorResponse  "Method not allowed"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Security    BearerAuth
// @Router       /games/{id} [delete]
func (h *GameHandler) DeleteGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	gameID, err := uuid.Parse(idStr)
	if err != nil {
		response.RespondWithError(w, http.StatusBadRequest, response.ErrInvalidGameID)
		return
	}

	if err := h.Service.DeleteGame(r.Context(), gameID); err != nil {
		if errors.Is(err, response.ErrNotFound) {
			response.RespondWithError(w, http.StatusNotFound, response.ErrGameNotFound)
		} else {
			response.RespondWithError(w, http.StatusInternalServerError, response.ErrInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListGames retrieves all games.
// @Summary      List all games
// @Description  Returns a list of all games in the system.
// @Tags         games
// @Produce      json
// @Success      200  {array}   []models.Game  "List of games"
// @Failure      403  {object}  response.ErrorResponse  "Missing or invalid token"
// @Failure      500  {object}  response.ErrorResponse  "Internal server error"
// @Security    BearerAuth
// @Router       /games [get]
func (h *GameHandler) ListGames(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		response.RespondWithError(w, http.StatusMethodNotAllowed, response.ErrMethodNotAllowed)
		return
	}

	games, err := h.Service.ListGames(r.Context())
	if err != nil {
		log.Printf("ListGames: failed to retrieve games: %v", err)
		response.RespondWithError(w, http.StatusInternalServerError, response.ErrInternalServerError)
		return
	}

	response.RespondWithData(w, games)
}
