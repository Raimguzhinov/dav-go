package caldav_db

import (
	"strconv"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
	"github.com/jackc/pgx/v5/pgtype"
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
}

func ScanEvent(event *ical.Component, wantSequence int) *Calendar {
	if event == nil {
		return nil
	}

	cal := Calendar{
		CompTypeBit:  pgtype.Text{Valid: false},
		Transparent:  pgtype.Text{Valid: false},
		AllDay:       pgtype.Text{String: "0", Valid: true},
		Summary:      getTextValue(event, ical.PropSummary),
		Description:  getTextValue(event, ical.PropDescription),
		Url:          getTextValue(event, ical.PropURL),
		Organizer:    getTextValue(event, ical.PropOrganizer),
		Class:        getTextValue(event, ical.PropClass),
		Loc:          getTextValue(event, ical.PropLocation),
		Status:       getTextValue(event, ical.PropStatus),
		Categories:   getTextValue(event, ical.PropCategories),
		Timestamp:    getTimeValue(event, ical.PropDateTimeStamp),
		Created:      getTimeValue(event, ical.PropCreated),
		LastModified: getTimeValue(event, ical.PropLastModified),
		Start:        getTimeValue(event, ical.PropDateTimeStart),
		End:          getTimeValue(event, ical.PropDateTimeEnd),
		Duration:     getIntValue(event, ical.PropDuration),
		Priority:     getIntValue(event, ical.PropPriority),
		Sequence:     updateSequence(event, wantSequence),
		Completed:    getIntValue(event, ical.PropCompleted),
		PerCompleted: getIntValue(event, ical.PropPercentComplete),
	}

	switch event.Name {
	case ical.CompEvent:
		cal.CompTypeBit = pgtype.Text{String: "1", Valid: true}
	case ical.CompToDo:
		cal.CompTypeBit = pgtype.Text{String: "0", Valid: true}
	}

	transparent := getTextValue(event, ical.PropTransparency)
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

func getTextValue(event *ical.Component, propName string) pgtype.Text {
	prop := event.Props.Get(propName)
	if prop == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: prop.Value, Valid: true}
}

func getIntValue(event *ical.Component, propName string, wantSequence ...int) pgtype.Uint32 {
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
	val, err := prop.Int()
	if err != nil {
		return pgtype.Uint32{Uint32: 1, Valid: true}
	}
	return pgtype.Uint32{Uint32: uint32(val + wantSequence), Valid: true}
}

func getTimeValue(event *ical.Component, propName string) pgtype.Timestamp {
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
