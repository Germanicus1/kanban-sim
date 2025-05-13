package models

import "github.com/google/uuid"

// EffortType defines the kinds of work tracked on each card.
type EffortType struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`      // e.g. "Analysis"
	OrderIndex int       `json:"orderIndex"` // to preserve configured order
}

// Effort is one slice of work on a card.
type Effort struct {
	EffortType string `json:"effortType"`          // e.g. "Development"
	Estimate   int    `json:"estimate"`            // 1–16 initial
	Remaining  *int   `json:"remaining,omitempty"` // from DB
	Actual     *int   `json:"actual,omitempty"`    // from DB
}
