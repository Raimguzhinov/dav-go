package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/jackc/pgx/v5/pgtype"
)

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
	RecurrenceSet       *RecurrenceSet   `json:"recurrenceSet,omitempty"`
	Properties          map[string]any   `json:"props,omitempty"`
	NotDeletedException string           `json:"notDeletedException,omitempty"`
}

func ScanEvent(event *ical.Component) *Event {
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
		Sequence:     intValue(event, ical.PropSequence),
		Completed:    intValue(event, ical.PropCompleted),
		PerCompleted: intValue(event, ical.PropPercentComplete),
		Properties:   make(map[string]any),
	}

	switch event.Name {
	case ical.CompEvent:
		e.CompTypeBit = BitIsSet
	case ical.CompToDo:
		e.CompTypeBit = BitNone
	}

	transparent := textValue(event, ical.PropTransparency)
	if transparent.Valid {
		switch transparent.String {
		case "OPAQUE":
			e.Transparent = BitIsSet
		case "TRANSPARENT":
			e.Transparent = BitNone
		}
	}

	for k, v := range event.Props {
		if strings.HasPrefix(k, "X-") {
			e.Properties[v[0].Name] = v[0].Value
		}
	}

	if e.End.Time.Sub(e.Start.Time) == time.Hour*24 {
		e.AllDay = BitIsSet
	}

	return &e
}

func (c *Event) ToDomain(uid string) *ical.Component {
	calEvent := ical.NewEvent()

	if c.CompTypeBit == BitIsSet {
		calEvent.Name = ical.CompEvent
	} else if c.CompTypeBit == BitNone {
		calEvent.Name = ical.CompToDo
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

	if c.Transparent == BitIsSet {
		setTextValue(calEvent, ical.PropTransparency, pgtype.Text{String: "OPAQUE", Valid: true})
	} else if c.Transparent == BitNone {
		setTextValue(calEvent, ical.PropTransparency, pgtype.Text{String: "TRANSPARENT", Valid: true})
	}

	for k, v := range c.Properties {
		custom := ical.NewProp(k)
		switch v.(type) {
		case string:
			custom.SetValueType(ical.ValueText)
			custom.Value = v.(string)
		case int:
			custom.SetValueType(ical.ValueInt)
			custom.Value = strconv.Itoa(v.(int))
		case float64:
			custom.SetValueType(ical.ValueFloat)
			custom.Value = strconv.FormatFloat(v.(float64), 'f', -1, 64)
		case time.Time:
			custom.SetValueType(ical.ValueDateTime)
			custom.Value = v.(time.Time).UTC().Format(datetimeUTCFormat)
		case bool:
			custom.SetValueType(ical.ValueBool)
			custom.Value = strconv.FormatBool(v.(bool))
		default:
			custom.SetValueType(ical.ValueDefault)
			custom.Value = v.(string)
		}
		calEvent.Props.Set(custom)
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
