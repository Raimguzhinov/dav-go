// Package repo implements repo connection.
package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	adapter "github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/jackc/pgx/v5"
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

	Pool  *pgxpool.Pool
	Batch *pgx.Batch
}

// New -.
func New(ctx context.Context, logger *adapter.Logger, url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  _defaultMaxPoolSize,
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,

		Batch: &pgx.Batch{},
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("repo - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)
	poolConfig.ConnConfig.Tracer = adapter.NewTracer(logger)

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
