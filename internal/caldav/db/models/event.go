package models

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/jackc/pgx/v5/pgtype"
)

type Event struct {
	CompTypeBit         pgtype.Text                       `json:"compTypeBit,omitempty"`
	Transparent         pgtype.Text                       `json:"transparent,omitempty"`
	AllDay              pgtype.Text                       `json:"allDay,omitempty"`
	Summary             pgtype.Text                       `json:"summary,omitempty"`
	Description         pgtype.Text                       `json:"description,omitempty"`
	Url                 pgtype.Text                       `json:"url,omitempty"`
	Organizer           pgtype.Text                       `json:"organizer,omitempty"`
	Class               pgtype.Text                       `json:"class,omitempty"`
	Loc                 pgtype.Text                       `json:"loc,omitempty"`
	Status              pgtype.Text                       `json:"status,omitempty"`
	Categories          pgtype.Text                       `json:"categories,omitempty"`
	Timestamp           pgtype.Timestamp                  `json:"timestamp,omitempty"`
	Created             pgtype.Timestamp                  `json:"created,omitempty"`
	LastModified        pgtype.Timestamp                  `json:"lastModified,omitempty"`
	Start               pgtype.Timestamp                  `json:"start,omitempty"`
	End                 pgtype.Timestamp                  `json:"end,omitempty"`
	Duration            pgtype.Uint32                     `json:"duration,omitempty"`
	Priority            pgtype.Uint32                     `json:"priority,omitempty"`
	Sequence            pgtype.Uint32                     `json:"sequence,omitempty"`
	Completed           pgtype.Uint32                     `json:"completed,omitempty"`
	PerCompleted        pgtype.Uint32                     `json:"perCompleted,omitempty"`
	RecurrenceSet       *RecurrenceSet                    `json:"recurrenceSet,omitempty"`
	Properties          map[string]map[ical.ValueType]any `json:"props,omitempty"`
	NotDeletedException string                            `json:"notDeletedException,omitempty"`
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
		Properties:   make(map[string]map[ical.ValueType]any),
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
			icalValue := v[0].Value
			typeValue := make(map[ical.ValueType]any)

			switch v[0].ValueType() {
			case ical.ValueText:
				typeValue[ical.ValueText] = icalValue
			case ical.ValueInt:
				if intVal, err := strconv.Atoi(icalValue); err == nil {
					val := intVal
					typeValue[ical.ValueInt] = val
				}
			case ical.ValueFloat:
				if floatVal, err := strconv.ParseFloat(icalValue, 64); err == nil {
					val := floatVal
					typeValue[ical.ValueFloat] = val
				}
			case ical.ValueBool:
				if boolVal, err := strconv.ParseBool(icalValue); err == nil {
					val := boolVal
					typeValue[ical.ValueBool] = val
				}
			case ical.ValueTime, ical.ValueDateTime, ical.ValueDate:
				if timeVal, err := time.Parse(datetimeUTCFormat, icalValue); err == nil {
					val := timeVal
					typeValue[ical.ValueDateTime] = val
				}
			case ical.ValueBinary:
				typeValue[ical.ValueBinary] = icalValue
			case ical.ValueDefault:
				var val any
				if timeVal, err := time.Parse(datetimeUTCFormat, icalValue); err == nil {
					val = timeVal
				} else if numVal, err := strconv.ParseFloat(icalValue, 64); err == nil {
					val = numVal
				} else if boolVal, err := strconv.ParseBool(icalValue); err == nil {
					val = boolVal
				} else {
					val = icalValue
				}
				typeValue[ical.ValueDefault] = val
			default:
				typeValue[v[0].ValueType()] = icalValue
			}

			e.Properties[v[0].Name] = typeValue
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

	for name, valueType := range c.Properties {
		custom := ical.NewProp(name)

		for typeName, icalValue := range valueType {
			switch typeName {
			case ical.ValueText:
				custom.SetValueType(ical.ValueText)
				custom.Value = icalValue.(string)
			case ical.ValueInt:
				custom.SetValueType(ical.ValueInt)
				custom.Value = strconv.FormatFloat(icalValue.(float64), 'f', -1, 64)
			case ical.ValueFloat:
				custom.SetValueType(ical.ValueFloat)
				custom.Value = strconv.FormatFloat(icalValue.(float64), 'f', -1, 64)
			case ical.ValueBool:
				custom.SetValueType(ical.ValueBool)
				custom.Value = strconv.FormatBool(icalValue.(bool))
			case ical.ValueTime, ical.ValueDateTime, ical.ValueDate:
				custom.SetValueType(ical.ValueDateTime)
				timeVal, _ := time.Parse(time.RFC3339, icalValue.(string))
				custom.Value = timeVal.UTC().Format(datetimeUTCFormat)
			case ical.ValueBinary:
				custom.SetValueType(ical.ValueBinary)
				custom.Value = icalValue.(string)
			case ical.ValueDefault:
				switch val := icalValue.(type) {
				case int:
					custom.SetValueType(ical.ValueInt)
					custom.Value = strconv.Itoa(val)
				case float64:
					custom.SetValueType(ical.ValueFloat)
					custom.Value = strconv.FormatFloat(val, 'f', -1, 64)
				case bool:
					custom.SetValueType(ical.ValueBool)
					custom.Value = strconv.FormatBool(val)
				case []byte:
					custom.SetValueType(ical.ValueBinary)
					custom.Value = string(val)
				default:
					isLetter := regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
					if timeVal, err := time.Parse(time.RFC3339, val.(string)); err == nil {
						custom.SetValueType(ical.ValueDateTime)
						custom.Value = timeVal.UTC().Format(datetimeUTCFormat)
						break
					} else if isLetter(val.(string)) {
						custom.SetValueType(ical.ValueText)
						custom.Value = val.(string)
						break
					}
					custom.SetValueType(ical.ValueDefault)
					custom.Value = val.(string)
				}
			default:
				custom.SetValueType(typeName)
				custom.Value = icalValue.(string)
			}
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
