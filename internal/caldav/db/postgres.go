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
	"github.com/jackc/pgx/v5/pgtype"
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
	Size  int64    `json:"size"`
}

func (r *repository) CreateFolder(
	ctx context.Context,
	homeSetPath string,
	calendar *caldav.Calendar,
) error {
	q := `
		INSERT INTO caldav.calendar_folder
			(name, description, types, max_size)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	r.logger.Debug(q)

	var f folder

	if err := r.client.Pool.QueryRow(ctx, q, calendar.Name, calendar.Description, calendar.SupportedComponentSet, calendar.MaxResourceSize).Scan(&f.ID); err != nil {
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
			array_agg(f.types) AS types,
			f.max_size AS size
		FROM
			caldav.calendar_folder f
		GROUP BY
			f.id, f.name, f.description
		ORDER BY
			f.id 
		`
	r.logger.Debug(q)

	rows, err := r.client.Pool.Query(ctx, q)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return nil, err
	}

	var f folder
	var calendars []caldav.Calendar

	for rows.Next() {
		var calendar caldav.Calendar

		err = rows.Scan(&f.ID, &calendar.Name, &calendar.Description, &f.Types, &f.Size)
		if err != nil {
			err = r.client.ToPgErr(err)
			r.logger.Error(err)
			return nil, err
		}

		calendar.Path = strconv.Itoa(f.ID)
		calendar.SupportedComponentSet = f.Types
		calendar.MaxResourceSize = f.Size

		calendars = append(calendars, calendar)
	}
	return calendars, nil
}

func (r *repository) GetFileInfo(ctx context.Context, uid string) (*caldav.CalendarObject, error) {
	q := `
		SELECT
		    etag, modified_at, size
		FROM
		    caldav.calendar_file
		WHERE
		    uid = $1
		`
	r.logger.Debug(q)

	var calendar caldav.CalendarObject

	if err := r.client.Pool.QueryRow(ctx, q, uid).Scan(
		&calendar.ETag, &calendar.ModTime, &calendar.ContentLength,
	); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
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
	q := `
		CALL caldav.create_or_update_calendar_file($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	r.logger.Debug(q)

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
		r.logger.Error(err)
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		q,
		uid,
		eventType,
		folderID,
		object.ETag,
		wantEtag,
		object.ModTime,
		object.ContentLength,
		version,
		prodID,
		ifNoneMatch,
		ifMatch,
	)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return nil, err
	}

	sq := `
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
	`
	r.logger.Info(sq)

	tz, _ := getTimezone(object)

	eg := errgroup.Group{}

	for _, child := range object.Data.Component.Children {
		if child.Name == ical.CompEvent || child.Name == ical.CompToDo {
			eg.Go(func() error {
				return r.createEvent(ctx, tx, tz, sq, uid, wantSeq, child)
			})
		}
	}

	if err = eg.Wait(); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return nil, err
	}
	return object, nil
}

func (r *repository) createEvent(
	ctx context.Context,
	tx *postgres.Tx,
	tz, query, uid string,
	wantSequence int,
	event *ical.Component,
) error {
	var compTypeBit string

	switch event.Name {
	case ical.CompEvent:
		compTypeBit = "1"
	case ical.CompToDo:
		compTypeBit = "0"
	default:
		return fmt.Errorf("unknown event: %s", event.Name)
	}

	location, _ := time.LoadLocation(tz)

	summary := getTextValue(event.Props.Get(ical.PropSummary))
	description := getTextValue(event.Props.Get(ical.PropDescription))
	organizer := getTextValue(event.Props.Get(ical.PropOrganizer))
	duration := getTextValue(event.Props.Get(ical.PropDuration))
	class := getTextValue(event.Props.Get(ical.PropClass))
	loc := getTextValue(event.Props.Get(ical.PropLocation))
	priority := getTextValue(event.Props.Get(ical.PropPriority))
	url := getTextValue(event.Props.Get(ical.PropURL))
	status := getTextValue(event.Props.Get(ical.PropStatus))
	categories := getTextValue(event.Props.Get(ical.PropCategories))
	//transp := getTextValue(event.Props.Get(ical.PropTransparency))
	completed := getTextValue(event.Props.Get(ical.PropCompleted))
	perCompleted := getTextValue(event.Props.Get(ical.PropPercentComplete))
	sequence := getTextValue(event.Props.Get(ical.PropSequence))
	if sequence == nil {
		sequence = new(string)
		*sequence = "1"
	} else {
		oldSeq, err := strconv.Atoi(*sequence)
		if err != nil {
			return err
		}
		*sequence = strconv.Itoa(oldSeq + wantSequence)
	}

	allDay := "0"

	start, err := event.Props.DateTime(ical.PropDateTimeStart, location)
	if err != nil {
		return err
	}
	end, err := event.Props.DateTime(ical.PropDateTimeEnd, location)
	if err != nil {
		return err
	}
	created, err := event.Props.DateTime(ical.PropCreated, location)
	if err != nil {
		return err
	}
	r.logger.Info(created.UTC().String())

	timestamp, err := event.Props.DateTime(ical.PropDateTimeStamp, location)
	if err != nil {
		return err
	}
	lastModified, err := event.Props.DateTime(ical.PropLastModified, location)
	if err != nil {
		return err
	}

	if end.Sub(start) == time.Hour*24 {
		allDay = "1"
	}

	_, err = tx.Exec(
		ctx,
		query,
		uid,
		compTypeBit,
		timestamp.UTC(),
		created.UTC(),
		lastModified.UTC(),
		summary,
		description,
		url,
		organizer,
		start.UTC(),
		end.UTC(),
		duration,
		allDay,
		class,
		loc,
		priority,
		sequence,
		status,
		categories,
		nil,
		completed,
		perCompleted,
	)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return err
	}
	return nil
}

func getTimezone(calendarObject *caldav.CalendarObject) (string, error) {
	for _, child := range calendarObject.Data.Component.Children {
		if child.Name == ical.CompTimezone {
			return child.Props.Text(ical.ParamTimezoneID)
		}
	}
	return time.UTC.String(), fmt.Errorf("timezone not found")
}

func getTextValue(prop *ical.Prop) *string {
	if prop == nil {
		return nil
	}
	return &prop.Value
}

func (r *repository) GetCalendar(
	ctx context.Context,
	uid string,
	propFilter []string,
) (*ical.Calendar, error) {
	q := `
		SELECT cp.version,
		       cp.product,
		       cp.scale,
		       cp.method,
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
	`
	r.logger.Debug(q)

	var (
		version, product                                                                    string
		compTypeBit, allDay, transp                                                         pgtype.Bool
		scale, method, summary, description, url, organizer, class, loc, status, categories pgtype.Text
		timestamp, created, lastModified, start, end                                        pgtype.Timestamp
		duration, priority, sequence, completed, perCompleted                               pgtype.Uint32
	)

	if err := r.client.Pool.QueryRow(ctx, q, uid).Scan(
		&version,
		&product,
		&scale,
		&method,
		&compTypeBit,
		&timestamp,
		&created,
		&lastModified,
		&summary,
		&description,
		&url,
		&organizer,
		&start,
		&end,
		&duration,
		&allDay,
		&class,
		&loc,
		&priority,
		&sequence,
		&status,
		&categories,
		&transp,
		&completed,
		&perCompleted,
	); err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
		return nil, err
	}

	calEvent := ical.NewEvent()

	if compTypeBit.Valid {
		if compTypeBit.Bool {
			calEvent.Name = ical.CompEvent
		} else {
			calEvent.Name = ical.CompToDo
		}
	}

	setTextValue(calEvent, ical.PropSummary, summary)
	setTextValue(calEvent, ical.PropDescription, description)
	setTextValue(calEvent, ical.PropUID, pgtype.Text{String: uid, Valid: true})
	setTextValue(calEvent, ical.PropOrganizer, organizer)
	setIntValue(calEvent, ical.PropDuration, duration)
	setTextValue(calEvent, ical.PropClass, class)
	setTextValue(calEvent, ical.PropLocation, loc)
	setIntValue(calEvent, ical.PropPriority, priority)
	setTextValue(calEvent, ical.PropURL, url)
	setIntValue(calEvent, ical.PropSequence, sequence)
	setTextValue(calEvent, ical.PropStatus, status)
	setTextValue(calEvent, ical.PropCategories, categories)
	//setTextValue(calEvent, ical.PropTransparency, transp)
	setIntValue(calEvent, ical.PropCompleted, completed)
	setIntValue(calEvent, ical.PropPercentComplete, perCompleted)
	setTimestampValue(calEvent, ical.PropDateTimeStart, start)
	setTimestampValue(calEvent, ical.PropDateTimeEnd, end)
	setTimestampValue(calEvent, ical.PropCreated, created)
	setTimestampValue(calEvent, ical.PropDateTimeStamp, timestamp)
	setTimestampValue(calEvent, ical.PropLastModified, lastModified)

	cal := ical.NewCalendar()

	cal.Props.SetText(ical.PropVersion, version)
	cal.Props.SetText(ical.PropProductID, product)
	if scale.Valid {
		cal.Props.SetText(ical.PropCalendarScale, scale.String)
	}
	if method.Valid {
		cal.Props.SetText(ical.PropMethod, method.String)
	}
	cal.Children = []*ical.Component{calEvent.Component}

	return cal, nil
}

func setTextValue(event *ical.Event, propName string, text pgtype.Text) {
	if text.Valid {
		event.Props.SetText(propName, text.String)
	}
}

func setIntValue(event *ical.Event, propName string, value pgtype.Uint32) {
	if value.Valid {
		event.Props.SetText(propName, strconv.Itoa(int(value.Uint32)))
	}
}

func setTimestampValue(event *ical.Event, propName string, value pgtype.Timestamp) {
	if value.Valid {
		event.Props.SetDateTime(propName, value.Time.UTC())
	}
}

func (r *repository) FindObjects(
	ctx context.Context,
	folderID int,
	propFilter []string,
) ([]caldav.CalendarObject, error) {
	q := `
		SELECT uid,
		       etag,
		       modified_at,
		       size
		FROM caldav.calendar_file
		WHERE calendar_folder_id = $1
    `
	r.logger.Debug(q)

	var result []caldav.CalendarObject

	rows, err := r.client.Pool.Query(ctx, q, &folderID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error(err)
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
			r.logger.Error(err)
			return nil, err
		}

		result = append(result, obj)
	}

	return result, nil
}
