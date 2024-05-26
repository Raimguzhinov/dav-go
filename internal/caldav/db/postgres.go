package caldav_db

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strconv"

	backend "github.com/Raimguhinov/dav-go/internal/caldav"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/jackc/pgconn"
)

type repository struct {
	client *postgres.Postgres
	logger *logger.Logger
}

func NewRepository(client *postgres.Postgres, logger *logger.Logger) backend.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}

type folder struct {
	ID    int      `json:"id"`
	Types []string `json:"types"`
}

func (r *repository) CreateFolder(ctx context.Context, homeSetPath string, calendar *caldav.Calendar) error {
	var f folder
	q := `
		SELECT caldav.create_calendar_folder($1, $2, $3, $4)
	`
	//for _, calendarType := range calendar.SupportedComponentSet {
	r.logger.Debug(q)
	if err := r.client.Pool.QueryRow(ctx, q, calendar.Name, calendar.SupportedComponentSet, calendar.Description, calendar.MaxResourceSize).Scan(&f.ID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Error(fmt.Errorf("repo error: %s, detail: %s, where: %s, code: %s, state: %v", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
		}
		return err
	}
	//}
	calendar.Path = path.Join(homeSetPath, strconv.Itoa(f.ID))
	return nil
}

func (r *repository) FindFolders(ctx context.Context) ([]caldav.Calendar, error) {
	q := `
		SELECT
			f.id,
			f.name,
			f.description,
			array_agg(p.namespace) AS types,
			max(CASE WHEN p.name = 'MaxResourceSize' THEN p.prop_value END) AS size
		FROM
			caldav.calendar_folder f
				JOIN
			caldav.calendar_folder_property p ON f.id = p.calendar_folder_id
		GROUP BY
			f.id, f.name, f.description
		ORDER BY
			f.id; 
		`
	r.logger.Debug(q)

	rows, err := r.client.Pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	var f folder
	calendars := make([]caldav.Calendar, 0)

	for rows.Next() {
		var calendar caldav.Calendar
		var maxResourceSize string

		err = rows.Scan(&f.ID, &calendar.Name, &calendar.Description, &f.Types, &maxResourceSize)
		if err != nil {
			return nil, err
		}

		if maxResourceSize != "" {
			size, err := strconv.Atoi(maxResourceSize)
			if err != nil {
				return nil, err
			}
			calendar.MaxResourceSize = int64(size)
		}

		calendar.Path = strconv.Itoa(f.ID)
		calendar.SupportedComponentSet = f.Types

		calendars = append(calendars, calendar)
	}
	return calendars, nil
}

func (r *repository) PutObject(ctx context.Context, uid, eventType string, object *caldav.CalendarObject, opts *caldav.PutCalendarObjectOptions) (string, error) {
	// calendar_folder_id тот, для которого calendar_folder.name = eventType
	q := `
		CALL caldav.create_or_update_calendar_file($1, $2, $3, $4, $5, $6, $7, $8)
	`
	r.logger.Debug(q)

	ifNoneMatch := opts.IfNoneMatch.IsWildcard()
	ifMatch := opts.IfMatch.IsSet()

	var (
		want string
		err  error
	)
	if ifMatch {
		want, err = opts.IfMatch.ETag()
		if err != nil {
			return "", webdav.NewHTTPError(http.StatusBadRequest, err)
		}
	}
	folderDir, _ := path.Split(object.Path)
	folderID := path.Base(folderDir)

	_, err = r.client.Pool.Exec(ctx, q, uid, eventType, folderID, object.ETag, want, object.ModTime, ifNoneMatch, ifMatch)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			r.logger.Error(fmt.Errorf("repo error: %s, detail: %s, where: %s, code: %s, state: %v", pgErr.Message, pgErr.Detail, pgErr.Where, pgErr.Code, pgErr.SQLState()))
		}
		return "", err
	}
	return object.Path, nil
}

func (r *repository) CreateEvent(ctx context.Context, calendar *caldav.Calendar) error {
	return nil
}
