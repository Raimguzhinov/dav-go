package models

import (
	"math"
	"regexp"
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

func toJSONFormat(icalValue string, icalType ical.ValueType) map[ical.ValueType]any {
	valueType := make(map[ical.ValueType]any)

	switch icalType {
	case ical.ValueText:
		valueType[ical.ValueText] = icalValue
	case ical.ValueInt:
		if intVal, err := strconv.Atoi(icalValue); err == nil {
			val := intVal
			valueType[ical.ValueInt] = val
		}
	case ical.ValueFloat:
		if floatVal, err := strconv.ParseFloat(icalValue, 64); err == nil {
			if isWholeNumber(floatVal) {
				val := int(floatVal)
				valueType[ical.ValueInt] = val
			} else {
				val := floatVal
				valueType[ical.ValueFloat] = val
			}
		}
	case ical.ValueBool:
		if boolVal, err := strconv.ParseBool(icalValue); err == nil {
			val := boolVal
			valueType[ical.ValueBool] = val
		}
	case ical.ValueTime, ical.ValueDateTime, ical.ValueDate:
		if timeVal, err := time.Parse(datetimeUTCFormat, icalValue); err == nil {
			val := timeVal
			valueType[ical.ValueDateTime] = val
		}
	case ical.ValueBinary:
		valueType[ical.ValueBinary] = icalValue
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
		valueType[ical.ValueDefault] = val
	default:
		valueType[icalType] = icalValue
	}
	return valueType
}

func fromJSONFormat(prop *ical.Prop, valueType map[ical.ValueType]any) {
	for icalType, icalValue := range valueType {
		switch icalType {
		case ical.ValueText:
			prop.SetValueType(ical.ValueText)
			prop.Value = icalValue.(string)
		case ical.ValueInt:
			prop.SetValueType(ical.ValueInt)
			prop.Value = strconv.FormatFloat(icalValue.(float64), 'f', -1, 64)
		case ical.ValueFloat:
			prop.SetValueType(ical.ValueFloat)
			prop.Value = strconv.FormatFloat(icalValue.(float64), 'f', -1, 64)
		case ical.ValueBool:
			prop.SetValueType(ical.ValueBool)
			prop.Value = strconv.FormatBool(icalValue.(bool))
		case ical.ValueTime, ical.ValueDateTime, ical.ValueDate:
			prop.SetValueType(ical.ValueDateTime)
			timeVal, _ := time.Parse(time.RFC3339, icalValue.(string))
			prop.Value = timeVal.UTC().Format(datetimeUTCFormat)
		case ical.ValueBinary:
			prop.SetValueType(ical.ValueBinary)
			prop.Value = icalValue.(string)
		case ical.ValueDefault:
			switch val := icalValue.(type) {
			case int:
				prop.SetValueType(ical.ValueInt)
				prop.Value = strconv.Itoa(val)
			case float64:
				prop.SetValueType(ical.ValueFloat)
				prop.Value = strconv.FormatFloat(val, 'f', -1, 64)
			case bool:
				prop.SetValueType(ical.ValueBool)
				prop.Value = strconv.FormatBool(val)
			case []byte:
				prop.SetValueType(ical.ValueBinary)
				prop.Value = string(val)
			default:
				isLetter := regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
				if timeVal, err := time.Parse(time.RFC3339, val.(string)); err == nil {
					prop.SetValueType(ical.ValueDateTime)
					prop.Value = timeVal.UTC().Format(datetimeUTCFormat)
					break
				} else if isLetter(val.(string)) {
					prop.SetValueType(ical.ValueText)
					prop.Value = val.(string)
					break
				}
				prop.SetValueType(ical.ValueDefault)
				prop.Value = val.(string)
			}
		default:
			prop.SetValueType(icalType)
			prop.Value = icalValue.(string)
		}
	}
}

func isWholeNumber(f float64) bool {
	_, frac := math.Modf(f)
	return frac == 0.0
}
