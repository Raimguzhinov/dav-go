package postgres

import (
	"github.com/jackc/pgx/v5"
)

type Batch struct {
	*pgx.Batch
}

func (p *Postgres) NewBatch() *Batch {
	return &Batch{Batch: &pgx.Batch{}}
}
