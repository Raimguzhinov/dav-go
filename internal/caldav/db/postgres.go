package db

import (
	"context"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"

	backend "github.com/Raimguhinov/dav-go/internal/caldav"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/caldav"
	"github.com/emersion/go-ical"
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

func (r *repository) CreateFolder(ctx context.Context, homeSetPath string, calendar *caldav.Calendar) error {
	r.logger.Debug("postgres.CreateFolder")

	var f Folder

	err := r.client.Pool.QueryRow(ctx, `
		INSERT INTO caldav.calendar_folder
			(name, description, types, max_size)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, calendar.Name, calendar.Description, calendar.SupportedComponentSet, calendar.MaxResourceSize).Scan(&f.ID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.CreateFolder", logger.Err(err))
		return err
	}
	calendar.Path = path.Join(homeSetPath, strconv.Itoa(f.ID))
	return nil
}

func (r *repository) FindFolders(ctx context.Context) ([]caldav.Calendar, error) {
	r.logger.Debug("postgres.FindFolders")

	rows, err := r.client.Pool.Query(ctx, `
		SELECT
			f.id,
			f.name,
			COALESCE(f.description, '') as description,
			array_agg(f.types) AS types,
			f.max_size AS size
		FROM
			caldav.calendar_folder f
		GROUP BY
			f.id, f.name, f.description
		ORDER BY
			f.id 
	`)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.FindFolders", logger.Err(err))
		return nil, err
	}

	var f Folder
	var calendars []caldav.Calendar

	for rows.Next() {
		err = rows.Scan(&f.ID, &f.Name, &f.Description, &f.Types, &f.Size)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.FindFolders", logger.Err(err))
			return nil, err
		}

		calendars = append(calendars, f.ToDomain())
	}
	return calendars, nil
}

func (r *repository) GetFileInfo(ctx context.Context, uid string) (*caldav.CalendarObject, error) {
	r.logger.Debug("postgres.GetFileInfo")

	var calendar caldav.CalendarObject

	if err := r.client.Pool.QueryRow(ctx, `
		SELECT
			etag, modified_at, size
		FROM
			caldav.calendar_file
		WHERE
			uid = $1
	`, uid).Scan(
		&calendar.ETag, &calendar.ModTime, &calendar.ContentLength,
	); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.GetFileInfo", logger.Err(err))
		return nil, err
	}

	return &calendar, nil
}

func (r *repository) PutObject(
	ctx context.Context,
	uid, eventType string,
	object *caldav.CalendarObject,
	opts *caldav.PutCalendarObjectOptions,
) (*caldav.CalendarObject, error) {
	r.logger.Debug("postgres.PutObject")

	ifNoneMatch := opts.IfNoneMatch.IsWildcard()
	ifMatch := opts.IfMatch.IsSet()

	var (
		wantEtag string
		wantSeq  int
		err      error
	)

	if ifMatch {
		wantEtag, err = opts.IfMatch.ETag()
		wantSeq++
		if err != nil {
			return nil, webdav.NewHTTPError(http.StatusBadRequest, err)
		}
	}
	folderDir, _ := path.Split(object.Path)
	folderID := path.Base(folderDir)
	version, err := object.Data.Component.Props.Text(ical.PropVersion)
	if err != nil {
		return nil, err
	}
	prodID, err := object.Data.Component.Props.Text(ical.PropProductID)
	if err != nil {
		return nil, err
	}

	tx, err := r.client.NewTx(ctx)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.PutObject", logger.Err(err))
		return nil, err
	}
	defer func(tx *postgres.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	_, err = tx.Exec(
		ctx, `
		CALL caldav.create_or_update_calendar_file($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, uid, eventType, folderID, object.ETag, wantEtag, object.ModTime, object.ContentLength,
		version, prodID, ifNoneMatch, ifMatch,
	)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.PutObject", logger.Err(err))
		return nil, err
	}

	eg := errgroup.Group{}

	for _, child := range object.Data.Component.Children {
		if child.Name == ical.CompEvent || child.Name == ical.CompToDo {
			eg.Go(func() error {
				return r.createEvent(ctx, tx, uid, wantSeq, child)
			})
		}
	}

	if err = eg.Wait(); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.PutObject", logger.Err(err))
		return nil, err
	}
	return object, nil
}

func (r *repository) createEvent(
	ctx context.Context,
	tx *postgres.Tx,
	uid string,
	wantSequence int,
	event *ical.Component,
) error {
	r.logger.Debug("postgres.createEvent")

	var parentID int

	cal := ScanEvent(event, wantSequence)

	err := tx.QueryRow(ctx, `
		INSERT INTO caldav.event_component
		(
			calendar_file_uid, 
			component_type,
			date_timestamp,
			created_at,
			last_modified_at,
			summary,
			description,
		 	url,
			organizer,
			start_date,
			end_date,
			duration,
			all_day,
			class,
			location,
			priority,
			sequence,
			status,
			categories,
			event_transparency,
			todo_completed,
			todo_percent_complete
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
		ON CONFLICT (calendar_file_uid, created_at) DO UPDATE SET
			component_type = EXCLUDED.component_type,
			date_timestamp = EXCLUDED.date_timestamp,
			last_modified_at = EXCLUDED.last_modified_at,
			summary = EXCLUDED.summary,
			description = EXCLUDED.description,
			url = EXCLUDED.url,
			organizer = EXCLUDED.organizer,
			start_date = EXCLUDED.start_date,
			end_date = EXCLUDED.end_date,
			duration = EXCLUDED.duration,
			all_day = EXCLUDED.all_day,
			class = EXCLUDED.class,
			location = EXCLUDED.location,
			priority = EXCLUDED.priority,
			sequence = EXCLUDED.sequence,
			status = EXCLUDED.status,
			categories = EXCLUDED.categories,
			event_transparency = EXCLUDED.event_transparency,
			todo_completed = EXCLUDED.todo_completed,
			todo_percent_complete = EXCLUDED.todo_percent_complete
		RETURNING id
	`, uid, cal.CompTypeBit,
		cal.Timestamp, cal.Created, cal.LastModified,
		cal.Summary, cal.Description, cal.Url, cal.Organizer,
		cal.Start, cal.End,
		cal.Duration, cal.AllDay, cal.Class, cal.Loc, cal.Priority,
		cal.Sequence, cal.Status, cal.Categories, cal.Transparent,
		cal.Completed, cal.PerCompleted,
	).Scan(&parentID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.createEvent", logger.Err(err))
		return err
	}

	var customProps []CustomProp
	for k, v := range event.Props {
		if strings.HasPrefix(k, "X-") {
			var custom CustomProp

			custom.ParentID = parentID
			custom.Name = v[0].Name
			custom.ParamName = string(v[0].ValueType())
			custom.Value = v[0].Value

			customProps = append(customProps, custom)
		}
	}
	r.logger.Debug("Scanned custom prop", slog.Any("prop", customProps))

	var batch = r.client.Batch

	for _, cp := range customProps {
		batch.Queue(`
			INSERT INTO caldav.custom_property
			(
				parent_id,
			 	calendar_file_uid,
				prop_name,
				parameter_name,
				value
			) VALUES ($1, $2, $3, $4, $5)
		`, cp.ParentID, uid, cp.Name, cp.ParamName, cp.Value)
	}

	if rs := ScanRecurrence(event); rs != nil {
		batch.Queue(`
			INSERT INTO caldav.recurrence
			(
				event_component_id,
				interval,
				until,
				count,
				week_start,
				by_day,
				by_month_day,
				by_month,
				period_day,
				by_set_pos,
				this_and_future
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, parentID, rs.Interval, rs.Until, rs.Cnt, rs.Wkst, rs.Weekdays,
			rs.Monthdays, rs.Months, rs.PeriodDay, rs.BySetPos, rs.ThisAndFuture,
		)
	}

	res := tx.SendBatch(ctx, batch)

	if err := res.Close(); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.createEvent", logger.Err(err))
		return err
	}

	return nil
}

func (r *repository) GetCalendar(
	ctx context.Context,
	uid string,
	propFilter []string,
) (*ical.Calendar, error) {
	r.logger.Debug("postgres.GetCalendar")

	var cal Calendar
	var eventID int

	if err := r.client.Pool.QueryRow(ctx, `
		SELECT
			cp.version,
			cp.product,
			cp.scale,
			cp.method,
			ec.id,
			ec.component_type,
			ec.date_timestamp,
			ec.created_at,
			ec.last_modified_at,
			ec.summary,
			ec.description,
			ec.url,
			ec.organizer,
			ec.start_date,
			ec.end_date,
			ec.duration,
			ec.all_day,
			ec.class,
			ec.location,
			ec.priority,
			ec.sequence,
			ec.status,
			ec.categories,
			ec.event_transparency,
			ec.todo_completed,
			ec.todo_percent_complete
		FROM caldav.calendar_property AS cp
		JOIN caldav.event_component AS ec
			ON cp.calendar_file_uid = ec.calendar_file_uid
		WHERE cp.calendar_file_uid = $1
	`, uid).Scan(
		&cal.Version, &cal.Product, &cal.Scale, &cal.Method,
		&eventID, &cal.CompTypeBit,
		&cal.Timestamp, &cal.Created, &cal.LastModified,
		&cal.Summary, &cal.Description, &cal.Url, &cal.Organizer,
		&cal.Start, &cal.End, &cal.Duration, &cal.AllDay, &cal.Class, &cal.Loc, &cal.Priority,
		&cal.Sequence, &cal.Status, &cal.Categories, &cal.Transparent, &cal.Completed, &cal.PerCompleted,
	); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.GetCalendar", logger.Err(err))
		return nil, err
	}

	rows, err := r.client.Pool.Query(ctx, `
		SELECT
			prop_name,
			parameter_name,
			value
		FROM
			caldav.custom_property
		WHERE
			calendar_file_uid = $1
			AND parent_id = $2
	`, uid, eventID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.GetCalendar", logger.Err(err))
		return nil, err
	}

	for rows.Next() {
		var prop CustomProp

		err = rows.Scan(
			&prop.Name,
			&prop.ParamName,
			&prop.Value,
		)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.GetCalendar", logger.Err(err))
			return nil, err
		}
		cal.CustomProps = append(cal.CustomProps, prop)
	}

	return cal.ToDomain(uid), nil
}

func (r *repository) FindObjects(
	ctx context.Context,
	folderID int,
	propFilter []string,
) ([]caldav.CalendarObject, error) {
	r.logger.Debug("postgres.FindObjects")

	var result []caldav.CalendarObject

	rows, err := r.client.Pool.Query(ctx, `
		SELECT
			uid,
			etag,
			modified_at,
			size
		FROM caldav.calendar_file
		WHERE calendar_folder_id = $1
	`, &folderID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.FindObjects", logger.Err(err))
		return nil, err
	}

	for rows.Next() {
		var obj caldav.CalendarObject

		err = rows.Scan(
			&obj.Path,
			&obj.ETag,
			&obj.ModTime,
			&obj.ContentLength,
		)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.FindObjects", logger.Err(err))
			return nil, err
		}

		result = append(result, obj)
	}

	return result, nil
}
