package internal

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

type Card struct {
	ID             string         `json:"id"`
	ClassOfService string         `json:"classOfService"`
	ColumnID       string         `json:"columnId"`
	ValueEstimate  string         `json:"valueEstimate"`
	Effort         EffortEstimate `json:"effort"`
	SelectedDay    *int           `json:"selectedDay,omitempty"`
	DeployedDay    *int           `json:"deployedDay,omitempty"`
}

type EffortEstimate struct {
	Analysis    int `json:"analysis"`
	Development int `json:"development"`
	Test        int `json:"test"`
}
