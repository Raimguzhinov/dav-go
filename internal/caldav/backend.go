package backend

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/Raimguzhinov/go-webdav/caldav"
	"github.com/emersion/go-ical"
)

type Backend struct {
	calendars []caldav.Calendar
	objectMap map[string][]caldav.CalendarObject
}

func New() *Backend {
	var propFindUserPrincipal = `
		<?xml version="1.0" encoding="UTF-8"?>
		<A:propfind xmlns:A="DAV:">
		  <A:prop>
		    <A:current-user-principal/>
		    <A:principal-URL/>
		    <A:resourcetype/>
		  </A:prop>
		</A:propfind>
	`
	b := &Backend{
		calendars: make([]caldav.Calendar, 0),
		objectMap: make(map[string][]caldav.CalendarObject),
	}
	b.calendars = append(b.calendars, caldav.Calendar{Path: "/user/calendars/a", SupportedComponentSet: []string{"VEVENT"}})
	b.calendars = append(b.calendars, caldav.Calendar{Path: "/user/calendars/b", SupportedComponentSet: []string{"VTODO"}})
	b.objectMap["/user/calendars/a"] = make([]caldav.CalendarObject, 0)
	b.objectMap["/user/calendars/b"] = make([]caldav.CalendarObject, 0)

	req := httptest.NewRequest("PROPFIND", "/user/calendars/", strings.NewReader(propFindUserPrincipal))
	req.Header.Set("Content-Type", "application/xml")

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
	return "/user/calendars/", nil
}

func (s *Backend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	return "/user/", nil
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

func (s *Backend) QueryCalendarObjects(ctx context.Context, path string, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	return nil, nil
}
