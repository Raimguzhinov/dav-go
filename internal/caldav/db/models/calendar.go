package models

import (
	"github.com/emersion/go-ical"
	"github.com/jackc/pgx/v5/pgtype"
)

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
