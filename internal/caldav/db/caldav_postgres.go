package db

import (
	"context"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"time"

	backend "github.com/Raimguzhinov/dav-go/internal/caldav"
	"github.com/Raimguzhinov/dav-go/internal/caldav/db/models"
	"github.com/Raimguzhinov/dav-go/pkg/logger"
	"github.com/Raimguzhinov/dav-go/pkg/postgres"
	"github.com/Raimguzhinov/dav-go/pkg/utils"
	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/caldav"
	"github.com/emersion/go-ical"
	"golang.org/x/sync/errgroup"
)

type repository struct {
	client *postgres.Postgres
	logger *logger.Logger
}

func NewRepository(client *postgres.Postgres, logger *logger.Logger) backend.RepositoryCaldav {
	return &repository{
		client: client,
		logger: logger,
	}
}

func (r *repository) CreateCalendar(ctx context.Context, homeSetPath string, calendar *caldav.Calendar) error {
	r.logger.Debug("postgres.CreateCalendar")

	var f models.Folder

	err := r.client.Pool.QueryRow(ctx, `
		INSERT INTO caldav.calendar_folder
			(name, description, types, max_size)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, calendar.Name, calendar.Description, calendar.SupportedComponentSet, calendar.MaxResourceSize).Scan(&f.ID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.CreateCalendar", logger.Err(err))
		return err
	}
	calendar.Path = path.Join(homeSetPath, strconv.Itoa(f.ID))
	return nil
}

func (r *repository) FindCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	r.logger.Debug("postgres.FindCalendars")

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
		r.logger.Error("postgres.FindCalendars", logger.Err(err))
		return nil, err
	}

	var f models.Folder
	var calendars []caldav.Calendar

	for rows.Next() {
		err = rows.Scan(&f.ID, &f.Name, &f.Description, &f.Types, &f.Size)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.FindCalendars", logger.Err(err))
			return nil, err
		}

		calendars = append(calendars, f.ToDomain())
	}
	return calendars, nil
}

func (r *repository) GetCalendarObjectInfo(ctx context.Context, uid string) (*caldav.CalendarObject, error) {
	r.logger.Debug("postgres.GetCalendarObjectInfo")

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
		r.logger.Error("postgres.GetCalendarObjectInfo", logger.Err(err))
		return nil, err
	}

	return &calendar, nil
}

func (r *repository) UpgradeCalendarObject(
	ctx context.Context,
	uid, eventType string,
	object *caldav.CalendarObject,
	opts *caldav.PutCalendarObjectOptions,
) (*caldav.CalendarObject, error) {
	r.logger.Debug("postgres.UpgradeCalendarObject")

	ifNoneMatch := opts.IfNoneMatch.IsWildcard()
	ifMatch := opts.IfMatch.IsSet()

	var wantEtag string
	var err error

	if ifMatch {
		wantEtag, err = opts.IfMatch.ETag()
		if err != nil {
			return nil, webdav.NewHTTPError(http.StatusBadRequest, err)
		}
	}

	var cal models.Calendar
	var f models.Folder

	folderDir, _ := path.Split(object.Path)
	f.ID, err = strconv.Atoi(path.Base(folderDir))
	if err != nil {
		return nil, err
	}

	cal.Version, err = object.Data.Component.Props.Text(ical.PropVersion)
	if err != nil {
		return nil, err
	}
	cal.Product, err = object.Data.Component.Props.Text(ical.PropProductID)
	if err != nil {
		return nil, err
	}

	tx, err := r.client.NewTx(ctx)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.UpgradeCalendarObject", logger.Err(err))
		return nil, err
	}
	defer func(tx *postgres.Tx, ctx context.Context) {
		_ = tx.Rollback(ctx)
	}(tx, ctx)

	_, err = tx.Exec(
		ctx, `
		CALL caldav.create_or_update_calendar_file($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, uid, eventType, f.ID, object.ETag, wantEtag, object.ModTime, object.ContentLength,
		cal.Version, cal.Product, ifNoneMatch, ifMatch,
	)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.UpgradeCalendarObject", logger.Err(err))
		return nil, err
	}

	eg := errgroup.Group{}
	batch := r.client.NewBatch()
	recurParent := utils.NewOnceValue()
	recurCnt := utils.NewOnceValue()
	recurCnt.Set(len(object.Data.Component.Children))

	for _, child := range object.Data.Component.Children {
		if child.Name == ical.CompEvent || child.Name == ical.CompToDo {
			eg.Go(func() error {
				return r.createEvent(ctx, tx, batch, uid, recurParent, recurCnt, child)
			})
		}
	}

	if err = eg.Wait(); err != nil {
		return nil, err
	}

	res := tx.SendBatch(ctx, batch.Batch)
	if err := res.Close(); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.createEvent send batch", logger.Err(err))
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.UpgradeCalendarObject", logger.Err(err))
		return nil, err
	}
	return object, nil
}

func (r *repository) createEvent(
	ctx context.Context,
	tx *postgres.Tx,
	batch *postgres.Batch,
	uid string,
	recurParent *utils.OnceValue,
	recurCnt *utils.OnceValue,
	event *ical.Component,
) error {
	r.logger.Debug("postgres.createEvent")

	var parentID int

	e := models.ScanEvent(event)

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
			todo_percent_complete,
			properties
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
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
			todo_percent_complete = EXCLUDED.todo_percent_complete,
			properties = EXCLUDED.properties
		RETURNING id
	`, uid, e.CompTypeBit,
		e.Timestamp, e.Created, e.LastModified,
		e.Summary, e.Description, e.Url, e.Organizer,
		e.Start, e.End,
		e.Duration, e.AllDay, e.Class, e.Loc, e.Priority,
		e.Sequence, e.Status, e.Categories, e.Transparent,
		e.Completed, e.PerCompleted, e.Properties,
	).Scan(&parentID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.createEvent", logger.Err(err))
		return err
	}

	shouldRemoveRecur := false
	if val := recurCnt.Get(); val != nil {
		cnt := val.(int)
		if cnt == 1 {
			r.logger.Debug("postgres.createEvent should remove recurrence", slog.Int("parentID", parentID))
			shouldRemoveRecur = true
		} else if cnt > 1 {
			rs, err := event.RecurrenceSet(time.UTC)
			if err != nil {
				return err
			}
			if rs != nil {
				var pgRS models.RecurrenceSet
				_, err := r.scanRecurrence(ctx, parentID, &pgRS)
				if err != nil {
					return err
				}
				modelRRuleString := rs.GetRRule().Options.RRuleString()
				pgRRule, _ := pgRS.ToDomain()
				pgRRuleString := pgRRule.String()
				if modelRRuleString != pgRRuleString {
					r.logger.Debug("postgres.createEvent should remove recurrence because recurrence changed", slog.Int("parentID", parentID))
					shouldRemoveRecur = true
				}
			}
		}
	}
	if shouldRemoveRecur {
		if err := r.removeRecurrence(ctx, tx, parentID); err != nil {
			return err
		}
	}

	if rs := models.ScanRecurrence(event); rs != nil {
		var recurrenceID int

		err := tx.QueryRow(ctx, `
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
			ON CONFLICT (event_component_id) DO UPDATE SET
				interval = EXCLUDED.interval,
				until = EXCLUDED.until,
				count = EXCLUDED.count,
				week_start = EXCLUDED.week_start,
				by_day = EXCLUDED.by_day,
				by_month_day = EXCLUDED.by_month_day,
				by_month = EXCLUDED.by_month,
				period_day = EXCLUDED.period_day,
				by_set_pos = EXCLUDED.by_set_pos,
				this_and_future = EXCLUDED.this_and_future
			RETURNING id
		`, parentID, rs.Interval, rs.Until, rs.Cnt, rs.Wkst, rs.Weekdays,
			rs.Monthdays, rs.Months, rs.PeriodDay, rs.BySetPos, rs.ThisAndFuture,
		).Scan(&recurrenceID)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.createEvent", logger.Err(err))
			return err
		}

		if rs.Exceptions != nil && !shouldRemoveRecur {
			for _, ex := range rs.Exceptions {
				batch.Queue(`
					INSERT INTO caldav.recurrence_exception
					(
						event_component_id,
						recurrence_id,
						exception_date,
						deleted_recurrence
					) VALUES ($1, $2, $3, $4)
					ON CONFLICT (recurrence_id, exception_date) DO UPDATE SET
				exception_date = EXCLUDED.exception_date,
				deleted_recurrence = EXCLUDED.deleted_recurrence
				`, parentID, recurrenceID, ex.Value, "1")
			}
		}
		recurParent.Set(recurrenceID)
	}

	if ex := models.ScanRecurrenceException(event); ex != nil {
		for {
			r.logger.Debug("getting recurrence id...")

			val := recurParent.Get()
			if val != nil {
				recurrenceID := val.(int)
				r.logger.Debug("got recurrence id", slog.Int("recurrence_id", recurrenceID))
				batch.Queue(`
					INSERT INTO caldav.recurrence_exception
					(
						event_component_id,
						recurrence_id,
						exception_date,
						deleted_recurrence
					) VALUES ($1, $2, $3, $4)
					ON CONFLICT (recurrence_id, exception_date) DO UPDATE SET
						exception_date = EXCLUDED.exception_date,
						deleted_recurrence = EXCLUDED.deleted_recurrence
				`, parentID, recurrenceID, ex.Value, "0")
				break
			}
		}
	}

	//var customProps []models.Prop
	//for k, v := range event.Properties {
	//	if strings.HasPrefix(k, "X-") {
	//		var custom models.Prop
	//
	//		custom.ParentID = parentID
	//		custom.Name = v[0].Name
	//		custom.ParamName = string(v[0].ValueType())
	//		custom.Value = v[0].Value
	//
	//		customProps = append(customProps, custom)
	//	}
	//}
	//r.logger.Debug("scanned custom prop", slog.Any("prop", customProps))
	//
	//for _, cp := range customProps {
	//	batch.Queue(`
	//		INSERT INTO caldav.custom_property
	//		(
	//		 	calendar_file_uid,
	//			event_component_id,
	//			prop_name,
	//			parameter_name,
	//			value
	//		) VALUES ($1, $2, $3, $4, $5)
	//		ON CONFLICT (calendar_file_uid, event_component_id, prop_name) DO UPDATE SET
	//			parameter_name = EXCLUDED.parameter_name,
	//			value = EXCLUDED.value
	//	`, uid, cp.ParentID, cp.Name, cp.ParamName, cp.Value)
	//}

	return nil
}

func (r *repository) removeRecurrence(ctx context.Context, tx *postgres.Tx, parentID int) error {
	var childEvents []int

	rows, err := tx.Query(ctx, `
				DELETE FROM caldav.recurrence_exception
				WHERE recurrence_id = (SELECT id FROM caldav.recurrence WHERE event_component_id = $1)
				RETURNING event_component_id
			`, parentID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.createEvent", logger.Err(err))
		return err
	}

	for rows.Next() {
		var childID int
		err := rows.Scan(&childID)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.createEvent", logger.Err(err))
			return err
		}
		if childID == parentID {
			continue
		}
		childEvents = append(childEvents, childID)
	}

	_, err = tx.Exec(ctx, `DELETE FROM caldav.recurrence WHERE event_component_id = $1`, parentID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.createEvent", logger.Err(err))
		return err
	}

	for _, childID := range childEvents {
		r.logger.Debug("postgres.createEvent delete recurrence", slog.Int("childID", childID))

		//_, err = tx.Exec(ctx, `DELETE FROM caldav.custom_property WHERE event_component_id = $1`, childID)
		//if err != nil {
		//	err = r.client.ToPgErr(err)
		//	r.logger.Error("postgres.createEvent", logger.Err(err))
		//	return err
		//}
		_, err = tx.Exec(ctx, `DELETE FROM caldav.event_component WHERE id = $1`, childID)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.createEvent", logger.Err(err))
			return err
		}
	}
	return nil
}

func (r *repository) GetCalendar(
	ctx context.Context,
	uid string,
	propFilter []string,
) (*ical.Calendar, error) {
	r.logger.Debug("postgres.GetCalendar")

	var cal models.Calendar
	isNotDeletedExceptions := make(map[int]string)

	if err := r.client.Pool.QueryRow(ctx, `
		SELECT
			version,
			product,
			scale,
			method
		FROM caldav.calendar_property
		WHERE calendar_file_uid = $1
	`, uid).Scan(&cal.Version, &cal.Product, &cal.Scale, &cal.Method); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.GetCalendar", logger.Err(err))
		return nil, err
	}

	rows, err := r.client.Pool.Query(ctx, `
		SELECT
			id,
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
			todo_percent_complete,
			properties
		FROM caldav.event_component
		WHERE calendar_file_uid = $1
	`, uid)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.GetCalendar", logger.Err(err))
		return nil, err
	}

	for rows.Next() {
		var event models.Event
		var eventID int

		if err := rows.Scan(
			&eventID, &event.CompTypeBit, &event.Timestamp, &event.Created, &event.LastModified,
			&event.Summary, &event.Description, &event.Url, &event.Organizer, &event.Start, &event.End,
			&event.Duration, &event.AllDay, &event.Class, &event.Loc, &event.Priority, &event.Sequence,
			&event.Status, &event.Categories, &event.Transparent, &event.Completed, &event.PerCompleted,
			&event.Properties,
		); err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.GetCalendar", logger.Err(err))
			return nil, err
		}

		//subrows, err := r.client.Pool.Query(ctx, `
		//	SELECT
		//		prop_name,
		//		parameter_name,
		//		value
		//	FROM
		//		caldav.custom_property
		//	WHERE
		//		calendar_file_uid = $1
		//		AND event_component_id = $2
		//`, uid, eventID)
		//if err != nil {
		//	err = r.client.ToPgErr(err)
		//	r.logger.Error("postgres.GetCalendar", logger.Err(err))
		//	return nil, err
		//}
		//
		//for subrows.Next() {
		//	var prop models.Prop
		//
		//	err = subrows.Scan(
		//		&prop.Name,
		//		&prop.ParamName,
		//		&prop.Value,
		//	)
		//	if err != nil {
		//		err = r.client.ToPgErr(err)
		//		r.logger.Error("postgres.GetCalendar", logger.Err(err))
		//		return nil, err
		//	}
		//	event.CustomProps = append(event.CustomProps, prop)
		//}

		var rs models.RecurrenceSet
		recurrenceID, err := r.scanRecurrence(ctx, eventID, &rs)
		if err != nil {
			return nil, err
		}

		subrows, err := r.client.Pool.Query(ctx, `
			SELECT
				event_component_id,
				exception_date,
				deleted_recurrence
			FROM
				caldav.recurrence_exception
			WHERE
				recurrence_id = $1
		`, recurrenceID)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.GetCalendar", logger.Err(err))
			return nil, err
		}

		for subrows.Next() {
			var ex models.RecurrenceException
			var exEventID int

			err = subrows.Scan(
				&exEventID, &ex.Value, &ex.IsDeleted,
			)
			if err != nil {
				err = r.client.ToPgErr(err)
				r.logger.Error("postgres.GetCalendar", logger.Err(err))
				return nil, err
			}

			if ex.IsDeleted == models.BitIsSet {
				rs.Exceptions = append(rs.Exceptions, &ex)
			} else if ex.IsDeleted == models.BitNone {
				isNotDeletedExceptions[exEventID] = ex.ToDomain()
			}
		}

		if _, ok := isNotDeletedExceptions[eventID]; ok {
			event.NotDeletedException = isNotDeletedExceptions[eventID]
		}

		event.RecurrenceSet = &rs
		cal.Events = append(cal.Events, event)
	}

	return cal.ToDomain(uid), nil
}

func (r *repository) scanRecurrence(ctx context.Context, eventID int, rs *models.RecurrenceSet) (int, error) {
	var recurrenceID int
	err := r.client.Pool.QueryRow(ctx, `
			SELECT
				id,
				interval,
				until,
				count,
				week_start,
				by_day,
				by_month_day,
				by_month,
				period_day,
				by_set_pos
			FROM
				caldav.recurrence
			WHERE
				event_component_id = $1
		`, eventID).Scan(
		&recurrenceID,
		&rs.Interval, &rs.Until, &rs.Cnt, &rs.Wkst, &rs.Weekdays,
		&rs.Monthdays, &rs.Months, &rs.PeriodDay, &rs.BySetPos,
	)
	if err != nil {
		if !r.client.IsNoRows(err) {
			err = r.client.ToPgErr(err)
			r.logger.Error("postgres.GetCalendar", logger.Err(err))
			return 0, err
		}
	}
	return recurrenceID, nil
}

func (r *repository) FindCalendarObjects(
	ctx context.Context,
	folderID int,
	propFilter []string,
) ([]caldav.CalendarObject, error) {
	r.logger.Debug("postgres.FindCalendarObjects")

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
		r.logger.Error("postgres.FindCalendarObjects", logger.Err(err))
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
			r.logger.Error("postgres.FindCalendarObjects", logger.Err(err))
			return nil, err
		}

		result = append(result, obj)
	}

	return result, nil
}
