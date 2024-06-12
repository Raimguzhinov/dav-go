package db

import (
	"strconv"
	"time"

	"github.com/ceres919/go-webdav/caldav"
	"github.com/emersion/go-ical"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/teambition/rrule-go"
)

type Folder struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Types       []string `json:"types"`
	Size        int64    `json:"size"`
}

func (f *Folder) ToDomain() caldav.Calendar {
	return caldav.Calendar{
		Path:                  strconv.Itoa(f.ID),
		Name:                  f.Name,
		Description:           f.Description,
		SupportedComponentSet: f.Types,
		MaxResourceSize:       f.Size,
	}
}

type CustomProp struct {
	ParentID  int    `json:"parentID"`
	Name      string `json:"name"`
	ParamName string `json:"paramName"`
	Value     string `json:"value"`
}

func (cp *CustomProp) ToDomain() *ical.Prop {
	custom := ical.NewProp(cp.Name)
	custom.SetValueType(ical.ValueType(cp.ParamName))
	custom.Value = cp.Value
	return custom
}

type RecurrenceSet struct {
	Interval      pgtype.Uint32     `json:"interval,omitempty"`
	Cnt           pgtype.Uint32     `json:"cnt,omitempty"`
	Until         pgtype.Date       `json:"until,omitempty"`
	Wkst          pgtype.Int2       `json:"wkst,omitempty"`
	BySetPos      pgtype.Array[int] `json:"bySetPos,omitempty"`
	Weekdays      pgtype.Int2       `json:"weekdays,omitempty"`
	Monthdays     pgtype.Uint32     `json:"monthdays,omitempty"`
	Months        pgtype.Int2       `json:"months,omitempty"`
	PeriodDay     pgtype.Int2       `json:"periodDay,omitempty"`
	ThisAndFuture pgtype.Text       `json:"thisAndFuture,omitempty"`
}

func ScanRecurrence(event *ical.Component) *RecurrenceSet {
	recurrenceSet, err := event.RecurrenceSet(time.UTC)
	if err != nil {
		return nil
	}

	if recurrenceSet == nil {
		return nil
	}

	rs := &RecurrenceSet{
		Interval:      pgtype.Uint32{Valid: false},
		Cnt:           pgtype.Uint32{Valid: false},
		Until:         pgtype.Date{Valid: false},
		Wkst:          pgtype.Int2{Valid: false},
		BySetPos:      pgtype.Array[int]{Valid: false},
		Weekdays:      pgtype.Int2{Valid: false},
		Monthdays:     pgtype.Uint32{Valid: false},
		Months:        pgtype.Int2{Valid: false},
		PeriodDay:     pgtype.Int2{Valid: false},
		ThisAndFuture: pgtype.Text{String: "1", Valid: true},
	}

	options := recurrenceSet.GetRRule().Options

	if options.Interval != 0 {
		rs.Interval = pgtype.Uint32{Uint32: uint32(options.Interval), Valid: true}
	}
	if options.Count != 0 {
		rs.Cnt = pgtype.Uint32{Uint32: uint32(options.Count), Valid: true}
	}
	if !options.Until.IsZero() {
		rs.Until = pgtype.Date{Time: options.Until, Valid: true}
		rs.ThisAndFuture = pgtype.Text{String: "0", Valid: true}
	}
	if options.Wkst.String() != "" {
		rs.Wkst = pgtype.Int2{Int16: int16(options.Wkst.Day()), Valid: true}
	}
	if options.Bysetpos != nil {
		rs.BySetPos = pgtype.Array[int]{Elements: options.Bysetpos, Valid: true}
	}

	weekdays, periodDay, months, monthdays := getMasks(&options)

	if weekdays != nil && *weekdays != 0 {
		rs.Weekdays = pgtype.Int2{Int16: int16(*weekdays), Valid: true}
	}
	if periodDay != nil && *periodDay != 0 {
		rs.PeriodDay = pgtype.Int2{Int16: int16(*periodDay), Valid: true}
	}
	if months != nil && *months != 0 {
		rs.Months = pgtype.Int2{Int16: int16(*months), Valid: true}
	}
	if monthdays != nil && *monthdays != 0 {
		rs.Monthdays = pgtype.Uint32{Uint32: uint32(*monthdays), Valid: true}
	}

	return rs
}

func getMasks(options *rrule.ROption) (*time.Weekday, *int, *time.Month, *int) {
	var weekdays time.Weekday
	var months time.Month
	var monthdays int
	var periodDay int

	if options.Byweekday != nil {
		for _, mask := range options.Byweekday {
			if mask.Day()+1 == 7 {
				weekdays |= 1 << time.Sunday
				continue
			}
			weekdays |= 1 << (mask.Day() + 1)
			periodDay = mask.N()
		}
	} else {
		switch options.Freq {
		case rrule.DAILY:
			weekdays = 127
		case rrule.WEEKLY:
			weekdays |= 1 << options.Dtstart.Weekday()
		default:
		}
	}

	if options.Bymonth != nil {
		for _, mask := range options.Bymonth {
			months |= 1 << mask
		}
	}

	if options.Bymonthday != nil {
		for _, mask := range options.Bymonthday {
			if mask < 1 || mask > 31 {
				monthdays |= 1 << 0
				continue
			}
			monthdays |= 1 << mask
		}
	}

	//for i := time.Sunday; i <= time.Saturday; i++ {
	//	if weekdays&time.Weekday(1<<i) != 0 {
	//		r.logger.Debug("weekdays mask",
	//			slog.String("weekdays", i.String()),
	//			slog.Int("every", periodDay),
	//		)
	//	}
	//}
	//r.logger.Debug("Scanned weekdays mask", slog.Any("mask", weekdays))
	//
	//for i := time.January; i <= time.December; i++ {
	//	if months&time.Month(1<<i) != 0 {
	//		r.logger.Debug("months mask", slog.String("months", i.String()))
	//	}
	//}
	//r.logger.Debug("Scanned months mask", slog.Any("mask", months))
	//
	//for i := 0; i <= 31; i++ {
	//	if monthdays&int(1<<i) != 0 {
	//		if i == 0 {
	//			r.logger.Debug("monthday mask", slog.String("monthday", "Last day of months"))
	//			continue
	//		}
	//		r.logger.Debug("monthday mask", slog.Int("monthday", i))
	//	}
	//}
	//r.logger.Debug("Scanned monthday mask", slog.Any("mask", monthdays))

	return &weekdays, &periodDay, &months, &monthdays
}

type Calendar struct {
	Version      string           `json:"version"`
	Product      string           `json:"product"`
	CompTypeBit  pgtype.Text      `json:"compTypeBit,omitempty"`
	Transparent  pgtype.Text      `json:"transparent,omitempty"`
	AllDay       pgtype.Text      `json:"allDay,omitempty"`
	Scale        pgtype.Text      `json:"scale,omitempty"`
	Method       pgtype.Text      `json:"method,omitempty"`
	Summary      pgtype.Text      `json:"summary,omitempty"`
	Description  pgtype.Text      `json:"description,omitempty"`
	Url          pgtype.Text      `json:"url,omitempty"`
	Organizer    pgtype.Text      `json:"organizer,omitempty"`
	Class        pgtype.Text      `json:"class,omitempty"`
	Loc          pgtype.Text      `json:"loc,omitempty"`
	Status       pgtype.Text      `json:"status,omitempty"`
	Categories   pgtype.Text      `json:"categories,omitempty"`
	Timestamp    pgtype.Timestamp `json:"timestamp,omitempty"`
	Created      pgtype.Timestamp `json:"created,omitempty"`
	LastModified pgtype.Timestamp `json:"lastModified,omitempty"`
	Start        pgtype.Timestamp `json:"start,omitempty"`
	End          pgtype.Timestamp `json:"end,omitempty"`
	Duration     pgtype.Uint32    `json:"duration,omitempty"`
	Priority     pgtype.Uint32    `json:"priority,omitempty"`
	Sequence     pgtype.Uint32    `json:"sequence,omitempty"`
	Completed    pgtype.Uint32    `json:"completed,omitempty"`
	PerCompleted pgtype.Uint32    `json:"perCompleted,omitempty"`
	CustomProps  []CustomProp     `json:"customProps,omitempty"`
}

func ScanEvent(event *ical.Component, wantSequence int) *Calendar {
	if event == nil {
		return nil
	}

	cal := Calendar{
		CompTypeBit:  pgtype.Text{Valid: false},
		Transparent:  pgtype.Text{Valid: false},
		AllDay:       pgtype.Text{String: "0", Valid: true},
		Summary:      textValue(event, ical.PropSummary),
		Description:  textValue(event, ical.PropDescription),
		Url:          textValue(event, ical.PropURL),
		Organizer:    textValue(event, ical.PropOrganizer),
		Class:        textValue(event, ical.PropClass),
		Loc:          textValue(event, ical.PropLocation),
		Status:       textValue(event, ical.PropStatus),
		Categories:   textValue(event, ical.PropCategories),
		Timestamp:    timeValue(event, ical.PropDateTimeStamp),
		Created:      timeValue(event, ical.PropCreated),
		LastModified: timeValue(event, ical.PropLastModified),
		Start:        timeValue(event, ical.PropDateTimeStart),
		End:          timeValue(event, ical.PropDateTimeEnd),
		Duration:     intValue(event, ical.PropDuration),
		Priority:     intValue(event, ical.PropPriority),
		Sequence:     updateSequence(event, wantSequence),
		Completed:    intValue(event, ical.PropCompleted),
		PerCompleted: intValue(event, ical.PropPercentComplete),
	}

	switch event.Name {
	case ical.CompEvent:
		cal.CompTypeBit = pgtype.Text{String: "1", Valid: true}
	case ical.CompToDo:
		cal.CompTypeBit = pgtype.Text{String: "0", Valid: true}
	}

	transparent := textValue(event, ical.PropTransparency)
	if transparent.Valid {
		switch transparent.String {
		case "OPAQUE":
			cal.Transparent = pgtype.Text{String: "1", Valid: true}
		case "TRANSPARENT":
			cal.Transparent = pgtype.Text{String: "0", Valid: true}
		}
	}

	if cal.End.Time.Sub(cal.Start.Time) == time.Hour*24 {
		cal.AllDay = pgtype.Text{String: "1", Valid: true}
	}

	return &cal
}

func textValue(event *ical.Component, propName string) pgtype.Text {
	prop := event.Props.Get(propName)
	if prop == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: prop.Value, Valid: true}
}

func intValue(event *ical.Component, propName string, wantSequence ...int) pgtype.Uint32 {
	prop := event.Props.Get(propName)
	if prop == nil {
		return pgtype.Uint32{Valid: false}
	}
	val, err := prop.Int()
	if err != nil {
		return pgtype.Uint32{Valid: false}
	}
	return pgtype.Uint32{Uint32: uint32(val), Valid: true}
}

func updateSequence(event *ical.Component, wantSequence int) pgtype.Uint32 {
	prop := event.Props.Get(ical.PropSequence)
	if prop == nil {
		return pgtype.Uint32{Uint32: 1, Valid: true}
	}
	val, err := strconv.Atoi(prop.Value)
	if err != nil {
		return pgtype.Uint32{Uint32: uint32(wantSequence), Valid: true}
	}
	return pgtype.Uint32{Uint32: uint32(val + wantSequence), Valid: true}
}

func timeValue(event *ical.Component, propName string) pgtype.Timestamp {
	prop := event.Props.Get(propName)
	if prop == nil {
		return pgtype.Timestamp{Valid: false}
	}
	val, err := prop.DateTime(time.UTC)
	if err != nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: val.UTC(), Valid: true}
}

func (c *Calendar) ToDomain(uid string) *ical.Calendar {
	calEvent := ical.NewEvent()

	if c.CompTypeBit.Valid {
		if c.CompTypeBit.String == "1" {
			calEvent.Name = ical.CompEvent
		} else {
			calEvent.Name = ical.CompToDo
		}
	}

	setTextValue(calEvent, ical.PropSummary, c.Summary)
	setTextValue(calEvent, ical.PropDescription, c.Description)
	setTextValue(calEvent, ical.PropUID, pgtype.Text{String: uid, Valid: true})
	setTextValue(calEvent, ical.PropOrganizer, c.Organizer)
	setIntValue(calEvent, ical.PropDuration, c.Duration)
	setTextValue(calEvent, ical.PropClass, c.Class)
	setTextValue(calEvent, ical.PropLocation, c.Loc)
	setIntValue(calEvent, ical.PropPriority, c.Priority)
	setTextValue(calEvent, ical.PropURL, c.Url)
	setIntValue(calEvent, ical.PropSequence, c.Sequence)
	setTextValue(calEvent, ical.PropStatus, c.Status)
	setTextValue(calEvent, ical.PropCategories, c.Categories)
	setIntValue(calEvent, ical.PropCompleted, c.Completed)
	setIntValue(calEvent, ical.PropPercentComplete, c.PerCompleted)
	setTimestampValue(calEvent, ical.PropDateTimeStart, c.Start)
	setTimestampValue(calEvent, ical.PropDateTimeEnd, c.End)
	setTimestampValue(calEvent, ical.PropCreated, c.Created)
	setTimestampValue(calEvent, ical.PropDateTimeStamp, c.Timestamp)
	setTimestampValue(calEvent, ical.PropLastModified, c.LastModified)

	if c.Transparent.Valid {
		if c.Transparent.String == "1" {
			setTextValue(calEvent, ical.PropTransparency, pgtype.Text{String: "OPAQUE", Valid: true})
		} else {
			setTextValue(calEvent, ical.PropTransparency, pgtype.Text{String: "TRANSPARENT", Valid: true})
		}
	}

	for _, custom := range c.CustomProps {
		calEvent.Props.Set(custom.ToDomain())
	}

	cal := ical.NewCalendar()

	cal.Props.SetText(ical.PropVersion, c.Version)
	cal.Props.SetText(ical.PropProductID, c.Product)
	if c.Scale.Valid {
		cal.Props.SetText(ical.PropCalendarScale, c.Scale.String)
	}
	if c.Method.Valid {
		cal.Props.SetText(ical.PropMethod, c.Method.String)
	}
	cal.Children = []*ical.Component{calEvent.Component}

	return cal
}

func setTextValue(event *ical.Event, propName string, text pgtype.Text) {
	if text.Valid {
		event.Props.SetText(propName, text.String)
	}
}

func setIntValue(event *ical.Event, propName string, value pgtype.Uint32) {
	if value.Valid {
		intProp := ical.NewProp(propName)
		intProp.SetValueType(ical.ValueInt)
		intProp.Value = strconv.Itoa(int(value.Uint32))
		event.Props.Set(intProp)
	}
}

func setTimestampValue(event *ical.Event, propName string, value pgtype.Timestamp) {
	if value.Valid {
		event.Props.SetDateTime(propName, value.Time.UTC())
	}
}
