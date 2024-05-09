package caldav

import (
	"context"
	"fmt"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav/caldav"
)

type Backend struct {
	repo Repository
}

func New(repository Repository) (*Backend, error) {
	b := &Backend{
		repo: repository,
	}

	if err := b.repo.CreateFolder(
		context.Background(),
		&caldav.Calendar{Name: "VCALENDAR", SupportedComponentSet: []string{"VEVENT"}},
	); err != nil {
		return nil, err
	}
	if err := b.repo.CreateFolder(
		context.Background(),
		&caldav.Calendar{Name: "VCALENDAR", SupportedComponentSet: []string{"VTODO"}},
	); err != nil {
		return nil, err
	}
	return b, nil
}

func (b *Backend) ListCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	return b.repo.FindFolders(ctx)
}

func (b *Backend) GetCalendar(ctx context.Context, path string) (*caldav.Calendar, error) {
	cals, err := b.repo.FindFolders(ctx)
	if err != nil {
		return nil, err
	}
	for _, cal := range cals {
		if cal.Path == path {
			return &cal, nil
		}
	}
	return nil, fmt.Errorf("calendar for path: %b not found", path)
}

func (b *Backend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	return "/admin/calendars/", nil
}

func (b *Backend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	return "/admin/", nil
}

func (b *Backend) DeleteCalendarObject(ctx context.Context, path string) error {
	//delete(b.objectMap, path)
	return nil
}

func (b *Backend) GetCalendarObject(ctx context.Context, path string, req *caldav.CalendarCompRequest) (*caldav.CalendarObject, error) {
	//for _, objs := range b.objectMap {
	//	for _, obj := range objs {
	//		if obj.Path == path {
	//			return &obj, nil
	//		}
	//	}
	//}
	return nil, fmt.Errorf("couldn't find calendar object at: %b", path)
}

func (b *Backend) PutCalendarObject(ctx context.Context, path string, calendar *ical.Calendar, opts *caldav.PutCalendarObjectOptions) (string, error) {
	//object := caldav.CalendarObject{
	//	Path: path,
	//	Data: calendar,
	//}
	//b.objectMap[path] = append(b.objectMap[path], object)
	return path, nil
}

func (b *Backend) ListCalendarObjects(ctx context.Context, path string, req *caldav.CalendarCompRequest) ([]caldav.CalendarObject, error) {
	return nil, nil
}

func (b *Backend) QueryCalendarObjects(ctx context.Context, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	return nil, nil
}
