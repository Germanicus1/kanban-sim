package models

type BoardConfig struct {
	Columns      []BoardColumn `json:"columns"`
	InitialCards []Card        `json:"initialCards"`
}

type BoardColumn struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	WipLimit   int           `json:"wipLimit,omitempty"`
	Subcolumns []BoardColumn `json:"subcolumns,omitempty"`
}
