package postgres

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Tx struct {
	tx pgx.Tx
	mu sync.Mutex
}

func (p *Postgres) NewTx(ctx context.Context) (*Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &Tx{tx: tx}, nil
}

func (ct *Tx) Rollback(ctx context.Context) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	return ct.tx.Rollback(ctx)
}

func (ct *Tx) Commit(ctx context.Context) error {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	return ct.tx.Commit(ctx)
}

func (ct *Tx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	return ct.tx.QueryRow(ctx, sql, args...)
}

func (ct *Tx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	return ct.tx.Exec(ctx, sql, args...)
}
