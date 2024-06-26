package models

import (
	"strconv"
	"time"

	"github.com/emersion/go-ical"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	datetimeUTCFormat = "20060102T150405Z"
)

var (
	BitNone  = pgtype.Text{String: "0", Valid: true}
	BitIsSet = pgtype.Text{String: "1", Valid: true}
)

func textValue(event *ical.Component, propName string) pgtype.Text {
	prop := event.Props.Get(propName)
	if prop == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: prop.Value, Valid: true}
}

func intValue(event *ical.Component, propName string) pgtype.Uint32 {
	prop := event.Props.Get(propName)
	if prop == nil {
		if propName == ical.PropSequence {
			return pgtype.Uint32{Uint32: 1, Valid: true}
		}
		return pgtype.Uint32{Valid: false}
	}
	val, err := prop.Int()
	if err != nil {
		if propName == ical.PropSequence {
			return pgtype.Uint32{Uint32: 1, Valid: true}
		}
		return pgtype.Uint32{Valid: false}
	}
	return pgtype.Uint32{Uint32: uint32(val), Valid: true}
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
