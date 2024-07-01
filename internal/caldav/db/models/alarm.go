package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Alarm struct {
	Action      pgtype.Text      `json:"action"`
	Trigger     pgtype.Timestamp `json:"trigger"`
	Summary     pgtype.Text      `json:"summary"`
	Description pgtype.Text      `json:"description"`
	Duration    pgtype.Uint32    `json:"duration"`
	Repeat      pgtype.Uint32    `json:"repeat"`
}
