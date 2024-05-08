package backend

import (
	"context"
	"fmt"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

type Backend struct {
	calendars []caldav.Calendar
	objectMap map[string][]caldav.CalendarObject
}

func New() *Backend {
	b := &Backend{
		calendars: make([]caldav.Calendar, 0),
		objectMap: make(map[string][]caldav.CalendarObject),
	}

	b.calendars = append(b.calendars, caldav.Calendar{Name: "VCALENDAR", Path: "/admin/calendars/a/", SupportedComponentSet: []string{"VEVENT"}})
	b.calendars = append(b.calendars, caldav.Calendar{Name: "VCALENDAR", Path: "/admin/calendars/b/", SupportedComponentSet: []string{"VTODO"}})

	return b
}

func (s *Backend) CreateCalendar(ctx context.Context, calendar *caldav.Calendar) error {
	return nil
}

func (s *Backend) ListCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	return s.calendars, nil
}

func (s *Backend) GetCalendar(ctx context.Context, path string) (*caldav.Calendar, error) {
	for _, cal := range s.calendars {
		if cal.Path == path {
			return &cal, nil
		}
	}
	return nil, fmt.Errorf("calendar for path: %s not found", path)
}

func (s *Backend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	return "/admin/calendars/", nil
}

func (s *Backend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	return "/admin/", nil
}

func (s *Backend) DeleteCalendarObject(ctx context.Context, path string) error {
	delete(s.objectMap, path)
	return nil
}

func (s *Backend) GetCalendarObject(ctx context.Context, path string, req *caldav.CalendarCompRequest) (*caldav.CalendarObject, error) {
	for _, objs := range s.objectMap {
		for _, obj := range objs {
			if obj.Path == path {
				return &obj, nil
			}
		}
	}
	return nil, fmt.Errorf("couldn't find calendar object at: %s", path)
}

func (s *Backend) PutCalendarObject(ctx context.Context, path string, calendar *ical.Calendar, opts *caldav.PutCalendarObjectOptions) (string, error) {
	object := caldav.CalendarObject{
		Path: path,
		Data: calendar,
	}
	s.objectMap[path] = append(s.objectMap[path], object)
	return path, nil
}

func (s *Backend) ListCalendarObjects(ctx context.Context, path string, req *caldav.CalendarCompRequest) ([]caldav.CalendarObject, error) {
	return s.objectMap[path], nil
}

func (s *Backend) QueryCalendarObjects(ctx context.Context, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	return nil, nil
}
