package model

import (
	"fmt"
	"time"
)

type PileDrivingRecordLine struct {
	ProjectId    int       `json:"project_id"`
	PileNumber   string    `json:"pile_number"`
	PileFieldId  int       `json:"pile_field_id"`
	StartDate    time.Time `json:"start_date"`
	FactPileHead int       `json:"fact_pile_head"`
	RecordedBy   string    `json:"recorded_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

func (p PileDrivingRecordLine) String() string {
	return fmt.Sprintf("{project_id=%d, pile_number=\"%s\", pile_field_id=%d, "+
		"start_date=%s, fact_pile_head=%d, recorded_by=\"%s\", "+
		"created_at=%s, updated_at=%s, status=%d}",
		p.ProjectId,
		p.PileNumber,
		p.PileFieldId,
		p.StartDate.Format(time.DateOnly),
		p.FactPileHead,
		p.RecordedBy,
		p.CreatedAt.Format(time.DateTime),
		p.UpdatedAt.Format(time.DateTime),
		p.Status)
}
