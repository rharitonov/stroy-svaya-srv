package model

import "time"

type PileDrivingRecordLine struct {
	ProjectId    int       `json:"project_id"`
	PileNumber   string    `json:"pile_number"`
	PileFieldId  int       `json:"pile_field_id"`
	StartDate    time.Time `json:"start_date"`
	FactPileHead int       `json:"fact_pile_head"`
	RecordedBy   string    `json:"recorded_by"`
	Status       int       `json:"status"` // 10 - planned, 20 - logged, 30 - all, 40 -approved
}

type PileFilter struct {
	ProjectId    int        `json:"project_id,omitempty"`
	PileNumber   *string    `json:"pile_number,omitempty"`
	PileFieldId  int        `json:"pile_field_id,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	FactPileHead *int       `json:"fact_pile_head,omitempty"`
	RecordedBy   *string    `json:"recorded_by,omitempty"`
	Status       int        `json:"status,omitempty"` // 10 - planned, 20 - logged, 30 - all, 40 -approved
}
