package internal

import (
	"errors"
	"net/http"

	"github.com/jackc/pgconn"
)

const (
	ErrGameNotFound     = "GAME_NOT_FOUND"
	ErrInvalidGameID    = "INVALID_GAME_ID"
	ErrDatabaseError    = "DATABASE_ERROR"
	ErrValidationFailed = "VALIDATION_FAILED"
	ErrDuplicateEntry   = "DUPLICATE_ENTRY"
	ErrInvalidInput     = "INVALID_INPUT"
	ErrCardNotFound     = "CARD_NOT_FOUND"
)

// MapPostgresError maps PostgreSQL error codes to HTTP status codes and error messages
func MapPostgresError(err error) (int, string) {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return http.StatusConflict, ErrDuplicateEntry
		case "23503":
			return http.StatusBadRequest, "INVALID_FOREIGN_KEY"
		case "23502":
			return http.StatusBadRequest, "MISSING_REQUIRED_FIELD"
		case "23514":
			return http.StatusBadRequest, "CHECK_CONSTRAINT_VIOLATION"
		case "22P02":
			return http.StatusBadRequest, ErrInvalidInput
		case "42601":
			return http.StatusInternalServerError, "SYNTAX_ERROR"
		case "42703":
			return http.StatusInternalServerError, "UNDEFINED_COLUMN"
		default:
			return http.StatusInternalServerError, ErrDatabaseError
		}
	}

	// Default case for generic SQL errors
	return http.StatusInternalServerError, ErrDatabaseError
}
