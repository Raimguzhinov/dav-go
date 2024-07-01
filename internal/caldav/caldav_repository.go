package caldav

import (
	"context"

	"github.com/ceres919/go-webdav/caldav"
	"github.com/emersion/go-ical"
)

type RepositoryCaldav interface {
	CreateCalendar(ctx context.Context, homeSetPath string, calendar *caldav.Calendar) error
	FindCalendars(ctx context.Context) ([]caldav.Calendar, error)
	GetCalendarObjectInfo(ctx context.Context, uid string) (*caldav.CalendarObject, error)
	UpgradeCalendarObject(ctx context.Context,
		uid, eventType string,
		object *caldav.CalendarObject,
		opts *caldav.PutCalendarObjectOptions,
	) (*caldav.CalendarObject, error)
	GetCalendar(ctx context.Context, uid string, propFilter []string) (*ical.Calendar, error)
	FindCalendarObjects(ctx context.Context, folderID int, propFilter []string) ([]caldav.CalendarObject, error)
}
