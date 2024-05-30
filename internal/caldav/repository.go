package caldav

import (
	"context"

	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

type Repository interface {
	CreateFolder(ctx context.Context, homeSetPath string, calendar *caldav.Calendar) error
	FindFolders(ctx context.Context) ([]caldav.Calendar, error)
	PutObject(
		ctx context.Context,
		uid, eventType string,
		object *caldav.CalendarObject,
		opts *caldav.PutCalendarObjectOptions,
	) (*caldav.CalendarObject, error)
	CreateEvent(
		ctx context.Context,
		tx *postgres.Tx,
		tz, query, uid string,
		event *ical.Component,
	) error
}
