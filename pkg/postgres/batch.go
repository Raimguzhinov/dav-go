package postgres

import (
	"sync"

	"github.com/jackc/pgx/v5"
)

type Batch struct {
	sync.Mutex
	*pgx.Batch
}

func (p *Postgres) NewBatch() *Batch {
	return &Batch{Batch: &pgx.Batch{}}
}

func (b *Batch) Queue(sql string, args ...any) *pgx.QueuedQuery {
	b.Lock()
	defer b.Unlock()
	return b.Batch.Queue(sql, args...)
}

func (b *Batch) Len() int {
	b.Lock()
	defer b.Unlock()
	return b.Batch.Len()
}
