// Package repo implements repo connection.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder squirrel.StatementBuilderType
	Pool    *pgxpool.Pool
}

// New -.
func New(ctx context.Context, url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("repo - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		ctx, cancel := context.WithTimeout(ctx, pg.connTimeout)
		defer cancel()

		pg.Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", pg.connAttempts)

		time.Sleep(pg.connTimeout)

		pg.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("repo - NewPostgres - connAttempts == 0: %w", err)
	}

	return pg, nil
}

// Close -.
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func (p *Postgres) ToPgErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return fmt.Errorf(
			"repo error: %s, detail: %s, where: %s, code: %s, state: %v",
			pgErr.Message,
			pgErr.Detail,
			pgErr.Where,
			pgErr.Code,
			pgErr.SQLState(),
		)
	}
	return err
}

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
