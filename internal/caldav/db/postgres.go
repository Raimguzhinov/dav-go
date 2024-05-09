package caldav_db

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	backend "github.com/Raimguhinov/dav-go/internal/caldav"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/emersion/go-webdav/caldav"
	"github.com/jackc/pgconn"
)

type repository struct {
	client *postgres.Postgres
	logger *logger.Logger
}

type folder struct {
	ID int `json:"id"`
}

func (r *repository) CreateFolder(ctx context.Context, calendar *caldav.Calendar) error {
	var f folder
	q := `
		INSERT INTO calendar_folder (name, description)
		VALUES ($1, $2)
		RETURNING id
	`
	r.logger.Debug(q)
	if err := r.client.Pool.QueryRow(ctx, q, calendar.Name, calendar.Description).Scan(&f.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Error(fmt.Errorf("repo error: %s, detail: %s, where: %s, code: %s, state: %v", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
		}
		return err
	}
	calendar.Path = "admin" + "/calendars/" + strconv.Itoa(f.ID) + "/"
	return nil
}

func (r *repository) FindFolders(ctx context.Context) ([]caldav.Calendar, error) {
	q := `
		SELECT id, name, description FROM calendar_folder
	`
	r.logger.Debug(q)

	rows, err := r.client.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	calendars := make([]caldav.Calendar, 0)

	for rows.Next() {
		var f folder
		var calendar caldav.Calendar

		err = rows.Scan(&f.ID, &calendar.Name, &calendar.Description)
		if err != nil {
			return nil, err
		}
		calendar.Path = "admin" + "/calendars/" + strconv.Itoa(f.ID) + "/"

		calendars = append(calendars, calendar)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return calendars, nil
}

func NewRepository(client *postgres.Postgres, logger *logger.Logger) backend.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}
