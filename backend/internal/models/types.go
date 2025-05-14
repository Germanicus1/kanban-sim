package models

type BoardConfig struct {
	EffortTypes []EffortType `json:"effortTypes"`
	Columns     []Column     `json:"columns"`
	Cards       []Card       `json:"cards"`
}

type BoardColumn struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	WipLimit   int           `json:"wipLimit,omitempty"`
	Subcolumns []BoardColumn `json:"subcolumns,omitempty"`
}
