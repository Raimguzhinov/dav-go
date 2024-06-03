package postgres

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

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
