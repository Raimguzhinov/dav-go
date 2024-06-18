package models

import (
	"fmt"
	"time"

	"github.com/emersion/go-ical"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/teambition/rrule-go"
)

type RecurrenceException struct {
	Value     pgtype.Timestamp `json:"value"`
	IsDeleted pgtype.Text      `json:"isDeleted"`
}

func ScanRecurrenceException(event *ical.Component) *RecurrenceException {
	exRecurrenceID := timeValue(event, ical.PropRecurrenceID)
	if exRecurrenceID.Valid {
		return &RecurrenceException{
			Value:     exRecurrenceID,
			IsDeleted: BitIsSet,
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
