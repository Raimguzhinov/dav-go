package caldav

import (
	"context"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

type Repository interface {
	CreateFolder(ctx context.Context, homeSetPath string, calendar *caldav.Calendar) error
	FindFolders(ctx context.Context) ([]caldav.Calendar, error)
	GetFileInfo(ctx context.Context, uid string) (*caldav.CalendarObject, error)
	PutObject(
		ctx context.Context,
		uid, eventType string,
		object *caldav.CalendarObject,
		opts *caldav.PutCalendarObjectOptions,
	) (*caldav.CalendarObject, error)
	GetCalendar(ctx context.Context, uid string, propFilter []string) (*ical.Calendar, error)
	FindObjects(
		ctx context.Context,
		folderID int,
		propFilter []string,
	) ([]caldav.CalendarObject, error)
}
