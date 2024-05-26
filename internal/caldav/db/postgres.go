package caldav_db

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	backend "github.com/Raimguhinov/dav-go/internal/caldav"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"golang.org/x/sync/errgroup"
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
	q := `
		SELECT caldav.create_calendar_folder($1, $2, $3, $4)
	`
	r.logger.Debug(q)

	var f folder

	if err := r.client.Pool.QueryRow(ctx, q, calendar.Name, calendar.SupportedComponentSet, calendar.Description, calendar.MaxResourceSize).Scan(&f.ID); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return err
	}
	calendar.Path = path.Join(homeSetPath, strconv.Itoa(f.ID))
	return nil
}

func (r *repository) FindFolders(ctx context.Context) ([]caldav.Calendar, error) {
	q := `
		SELECT
			f.id,
			f.name,
    		COALESCE(f.description, '') as description,
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
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return nil, err
	}

	var f folder
	calendars := make([]caldav.Calendar, 0)

	for rows.Next() {
		var calendar caldav.Calendar
		var maxResourceSize string

		err = rows.Scan(&f.ID, &calendar.Name, &calendar.Description, &f.Types, &maxResourceSize)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error(err)
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
	q := `
		CALL caldav.create_or_update_calendar_file($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
	version, err := object.Data.Component.Props.Text(ical.PropVersion)
	if err != nil {
		return "", err
	}
	prodID, err := object.Data.Component.Props.Text(ical.PropProductID)
	if err != nil {
		return "", err
	}

	tx, err := r.client.NewTx(ctx)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return "", err
	}
	defer tx.Rollback(ctx)

	_, err = r.client.Pool.Exec(
		ctx, q, uid, eventType, folderID, object.ETag, want, object.ModTime, version, prodID, ifNoneMatch, ifMatch,
	)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return "", err
	}

	sq := `
		INSERT INTO caldav.event_component
		(
			calendar_file_uid, 
			component_type,
			date_timestamp
		) VALUES ($1, $2, $3)
	`
	//-- 			created_at,
	//-- 			last_modified_at,
	//-- 			summary,
	//-- 			description,
	//-- 			organizer_email,
	//-- 			organizer_common_name,
	//-- 			start_date,
	//-- 			start_timezone_id,
	//-- 			end_date,
	//-- 			end_timezone_id,
	//-- 			duration,
	//-- 			all_day,
	//-- 			class,
	//-- 			location,
	//-- 			priority,
	//-- 			sequence,
	//-- 			status,
	//-- 			categories,
	//-- 			event_transparency,
	//-- 			todo_completed,
	//-- 			todo_percent_complete
	//--, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)
	r.logger.Info(sq)

	eg := errgroup.Group{}
	for _, event := range object.Data.Component.Children {
		eg.Go(func() error {
			var compTypeBit string

			if event.Name == ical.CompEvent {
				compTypeBit = "1"
			} else if event.Name == ical.CompToDo {
				compTypeBit = "0"
			} else {
				return fmt.Errorf("unknown event: %s", event.Name)
			}

			_, err = r.client.Pool.Exec(
				ctx, sq, uid, compTypeBit, time.Now().UTC(),
			)
			if err != nil {
				err = r.client.ToPgErr(err)
				r.logger.Error(err)
				return err
			}
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		return "", err
	}

	if err = tx.Commit(ctx); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return "", err
	}
	return object.Path, nil
}

func (r *repository) CreateEvent(ctx context.Context, calendar *caldav.Calendar) error {
	return nil
}
