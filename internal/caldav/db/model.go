package db

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ceres919/go-webdav/caldav"
	"github.com/emersion/go-ical"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/teambition/rrule-go"
)

var (
	BitNone = pgtype.Text{String: "0", Valid: true}
	BitTrue = pgtype.Text{String: "1", Valid: true}
)

const (
	datetimeUTCFormat = "20060102T150405Z"
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

type RecurrenceException struct {
	Value     pgtype.Timestamp `json:"value"`
	IsDeleted pgtype.Text      `json:"isDeleted"`
}

func ScanRecurrenceException(event *ical.Component) *RecurrenceException {
	exRecurrenceID := timeValue(event, ical.PropRecurrenceID)
	if exRecurrenceID.Valid {
		return &RecurrenceException{
			Value:     exRecurrenceID,
			IsDeleted: BitTrue,
		}
	}
	return nil
}

func (r *RecurrenceException) ToDomain() string {
	return r.Value.Time.UTC().Format(datetimeUTCFormat)
}

type RecurrenceSet struct {
	Interval      pgtype.Uint32          `json:"interval,omitempty"`
	Cnt           pgtype.Uint32          `json:"cnt,omitempty"`
	Until         pgtype.Date            `json:"until,omitempty"`
	Wkst          pgtype.Uint32          `json:"wkst,omitempty"`
	BySetPos      pgtype.Array[int]      `json:"bySetPos,omitempty"`
	Weekdays      pgtype.Uint32          `json:"weekdays,omitempty"`
	Monthdays     pgtype.Uint32          `json:"monthdays,omitempty"`
	Months        pgtype.Uint32          `json:"months,omitempty"`
	PeriodDay     *int                   `json:"periodDay,omitempty"`
	ThisAndFuture pgtype.Text            `json:"thisAndFuture,omitempty"`
	Exceptions    []*RecurrenceException `json:"exceptions,omitempty"`
}

func ScanRecurrence(event *ical.Component) *RecurrenceSet {
	recurrenceSet, err := event.RecurrenceSet(time.UTC)
	if err != nil {
		return nil
	}
	if recurrenceSet == nil {
		return nil
	}

	standardDay := map[rrule.Weekday]time.Weekday{
		rrule.SU: time.Sunday,
		rrule.MO: time.Monday,
		rrule.TU: time.Tuesday,
		rrule.WE: time.Wednesday,
		rrule.TH: time.Thursday,
		rrule.FR: time.Friday,
		rrule.SA: time.Saturday,
	}

	rs := &RecurrenceSet{
		Interval:      pgtype.Uint32{Valid: false},
		Cnt:           pgtype.Uint32{Valid: false},
		Until:         pgtype.Date{Valid: false},
		Wkst:          pgtype.Uint32{Valid: false},
		BySetPos:      pgtype.Array[int]{Valid: false},
		Weekdays:      pgtype.Uint32{Valid: false},
		Monthdays:     pgtype.Uint32{Valid: false},
		Months:        pgtype.Uint32{Valid: false},
		ThisAndFuture: pgtype.Text{String: "1", Valid: true},
	}

	if recurrenceSet.GetExDate() != nil {
		rs.Exceptions = make([]*RecurrenceException, len(recurrenceSet.GetExDate()))
		for i, exDate := range recurrenceSet.GetExDate() {
			rs.Exceptions[i] = &RecurrenceException{
				Value: pgtype.Timestamp{Time: exDate, Valid: true},
			}
		}
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
		rs.Wkst = pgtype.Uint32{Uint32: uint32(standardDay[options.Wkst]), Valid: true}
	}
	if options.Bysetpos != nil {
		rs.BySetPos = pgtype.Array[int]{Elements: options.Bysetpos, Valid: true}
	}

	weekdays, periodDay, months, monthdays := getMasks(&options, standardDay)

	if weekdays != nil && *weekdays != 0 {
		rs.Weekdays = pgtype.Uint32{Uint32: uint32(*weekdays), Valid: true}
	}
	if periodDay != nil && *periodDay != 0 {
		rs.PeriodDay = periodDay
	}
	if months != nil && *months != 0 {
		rs.Months = pgtype.Uint32{Uint32: uint32(*months), Valid: true}
	}
	if monthdays != nil && *monthdays != 0 {
		rs.Monthdays = pgtype.Uint32{Uint32: uint32(*monthdays), Valid: true}
	}

	return rs
}

func (rs *RecurrenceSet) ToDomain() (*rrule.ROption, string) {
	ro := rrule.ROption{Freq: rrule.SECONDLY}

	rruleDay := map[time.Weekday]rrule.Weekday{
		time.Sunday:    rrule.SU,
		time.Monday:    rrule.MO,
		time.Tuesday:   rrule.TU,
		time.Wednesday: rrule.WE,
		time.Thursday:  rrule.TH,
		time.Friday:    rrule.FR,
		time.Saturday:  rrule.SA,
	}

	if rs.Interval.Valid {
		ro.Interval = int(rs.Interval.Uint32)
	}
	if rs.Cnt.Valid {
		ro.Count = int(rs.Cnt.Uint32)
	}
	if rs.Until.Valid {
		ro.Until = rs.Until.Time
	}
	if rs.BySetPos.Valid {
		ro.Bysetpos = rs.BySetPos.Elements
	}
	if rs.Wkst.Valid {
		ro.Wkst = rruleDay[time.Weekday(rs.Wkst.Uint32)]
	}
	if rs.Weekdays.Valid {
		if rs.Weekdays.Uint32 == 127 {
			ro.Freq = rrule.DAILY
		} else {
			ro.Freq = rrule.WEEKLY
			for i := time.Sunday; i <= time.Saturday; i++ {
				if rs.Weekdays.Uint32&uint32(time.Weekday(1<<i)) != 0 {
					weekday := rruleDay[i]
					if rs.PeriodDay == nil {
						ro.Byweekday = append(ro.Byweekday, weekday)
						continue
					}
					ro.Byweekday = append(ro.Byweekday, weekday.Nth(*rs.PeriodDay))
				}
			}
		}
	}
	if rs.Monthdays.Valid {
		ro.Freq |= rrule.MONTHLY
		for i := 0; i <= 31; i++ {
			if rs.Monthdays.Uint32&uint32(1<<i) != 0 {
				if i == 0 {
					ro.Bymonthday = append(ro.Bymonthday, -1)
					continue
				}
				ro.Bymonthday = append(ro.Bymonthday, i)
			}
		}
	}
	if rs.Months.Valid {
		ro.Freq |= rrule.YEARLY
		for i := time.January; i <= time.December; i++ {
			if rs.Months.Uint32&uint32(time.Month(1<<i)) != 0 {
				ro.Bymonth = append(ro.Bymonth, int(i))
			}
		}
	}

	var exString string
	for i, exception := range rs.Exceptions {
		if exception.Value.Valid {
			if i == 0 {
				exString = exception.ToDomain()
				continue
			}
			exString = fmt.Sprintf("%s,%s", exString, exception.ToDomain())
		}
	}

	if ro.Freq == rrule.SECONDLY {
		return nil, exString
	}
	return &ro, exString
}

type Calendar struct {
	Version string      `json:"version"`
	Product string      `json:"product"`
	Scale   pgtype.Text `json:"scale,omitempty"`
	Method  pgtype.Text `json:"method,omitempty"`
	Events  []Event     `json:"events"`
}

func (c *Calendar) ToDomain(uid string) *ical.Calendar {
	cal := ical.NewCalendar()

	cal.Props.SetText(ical.PropVersion, c.Version)
	cal.Props.SetText(ical.PropProductID, c.Product)
	if c.Scale.Valid {
		cal.Props.SetText(ical.PropCalendarScale, c.Scale.String)
	}
	if c.Method.Valid {
		cal.Props.SetText(ical.PropMethod, c.Method.String)
	}
	cal.Children = make([]*ical.Component, 0, len(c.Events))
	for _, event := range c.Events {
		cal.Children = append(cal.Children, event.ToDomain(uid))
	}

	return cal
}

type Event struct {
	CompTypeBit         pgtype.Text      `json:"compTypeBit,omitempty"`
	Transparent         pgtype.Text      `json:"transparent,omitempty"`
	AllDay              pgtype.Text      `json:"allDay,omitempty"`
	Summary             pgtype.Text      `json:"summary,omitempty"`
	Description         pgtype.Text      `json:"description,omitempty"`
	Url                 pgtype.Text      `json:"url,omitempty"`
	Organizer           pgtype.Text      `json:"organizer,omitempty"`
	Class               pgtype.Text      `json:"class,omitempty"`
	Loc                 pgtype.Text      `json:"loc,omitempty"`
	Status              pgtype.Text      `json:"status,omitempty"`
	Categories          pgtype.Text      `json:"categories,omitempty"`
	Timestamp           pgtype.Timestamp `json:"timestamp,omitempty"`
	Created             pgtype.Timestamp `json:"created,omitempty"`
	LastModified        pgtype.Timestamp `json:"lastModified,omitempty"`
	Start               pgtype.Timestamp `json:"start,omitempty"`
	End                 pgtype.Timestamp `json:"end,omitempty"`
	Duration            pgtype.Uint32    `json:"duration,omitempty"`
	Priority            pgtype.Uint32    `json:"priority,omitempty"`
	Sequence            pgtype.Uint32    `json:"sequence,omitempty"`
	Completed           pgtype.Uint32    `json:"completed,omitempty"`
	PerCompleted        pgtype.Uint32    `json:"perCompleted,omitempty"`
	CustomProps         []CustomProp     `json:"customProps,omitempty"`
	RecurrenceSet       *RecurrenceSet   `json:"recurrenceSet,omitempty"`
	NotDeletedException string           `json:"notDeletedException,omitempty"`
}

func ScanEvent(event *ical.Component, wantSequence int) *Event {
	if event == nil {
		return nil
	}

	e := Event{
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
		e.CompTypeBit = pgtype.Text{String: "1", Valid: true}
	case ical.CompToDo:
		e.CompTypeBit = pgtype.Text{String: "0", Valid: true}
	}

	transparent := textValue(event, ical.PropTransparency)
	if transparent.Valid {
		switch transparent.String {
		case "OPAQUE":
			e.Transparent = pgtype.Text{String: "1", Valid: true}
		case "TRANSPARENT":
			e.Transparent = pgtype.Text{String: "0", Valid: true}
		}
	}

	if e.End.Time.Sub(e.Start.Time) == time.Hour*24 {
		e.AllDay = pgtype.Text{String: "1", Valid: true}
	}

	return &e
}

func (c *Event) ToDomain(uid string) *ical.Component {
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

	rs, exString := c.RecurrenceSet.ToDomain()
	if rs != nil {
		rs.Dtstart = c.Start.Time.UTC()
		calEvent.Props.SetRecurrenceRule(rs)
	}
	if exString != "" {
		exProp := ical.NewProp(ical.PropExceptionDates)
		exProp.SetValueType(ical.ValueDateTime)
		exProp.Value = exString
		calEvent.Props.Set(exProp)
	}
	if c.NotDeletedException != "" {
		recurrenceIDProp := ical.NewProp(ical.PropRecurrenceID)
		recurrenceIDProp.SetValueType(ical.ValueDateTime)
		recurrenceIDProp.Value = c.NotDeletedException
		calEvent.Props.Set(recurrenceIDProp)
	}

	return calEvent.Component
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

func getMasks(options *rrule.ROption, standardDay map[rrule.Weekday]time.Weekday) (*time.Weekday, *int, *time.Month, *int) {
	var weekdays time.Weekday
	var months time.Month
	var monthdays int
	var periodDay int

	if options.Byweekday != nil {
		for _, mask := range options.Byweekday {
			weekdays |= 1 << standardDay[mask]
			periodDay = mask.N()
		}
	} else {
		switch options.Freq {
		case rrule.DAILY:
			weekdays = 127
		case rrule.WEEKLY:
			weekdays |= 1 << options.Dtstart.UTC().Weekday()
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

	return &weekdays, &periodDay, &months, &monthdays
}
