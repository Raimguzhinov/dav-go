package caldav

import (
	"context"

	"github.com/emersion/go-webdav/caldav"
)

type Repository interface {
	CreateFolder(ctx context.Context, homeSetPath string, calendar *caldav.Calendar) error
	FindFolders(ctx context.Context) ([]caldav.Calendar, error)
	PutObject(ctx context.Context, uid, eventType string, object *caldav.CalendarObject, opts *caldav.PutCalendarObjectOptions) (string, error)
	CreateEvent(ctx context.Context, calendar *caldav.Calendar) error
}
