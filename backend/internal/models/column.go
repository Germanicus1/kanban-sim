package models

import "github.com/google/uuid"

// Column represents one column (or sub-column) on the board.
// You can nest SubColumns by building a tree from ParentID.
type Column struct {
	ID         uuid.UUID  `json:"id"`
	ParentID   *uuid.UUID `json:"parentId,omitempty"` // nil for top-level
	Title      string     `json:"title"`
	OrderIndex int        `json:"orderIndex"`
	WIPLimit   int        `json:"wipLimit,omitempty"`   // only set if non-zero
	Type       string     `json:"type"`                 // "active", queue", "done"
	SubColumns []Column   `json:"subColumns,omitempty"` // built in memory
}
