package caldav

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Raimguhinov/dav-go/internal/usecase/etag"
	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

type backend struct {
	webdav.UserPrincipalBackend
	prefix string
	repo   Repository
}

func New(
	upBackend webdav.UserPrincipalBackend,
	prefix string,
	repository Repository,
) (caldav.Backend, error) {
	b := &backend{
		UserPrincipalBackend: upBackend,
		prefix:               prefix,
		repo:                 repository,
	}
	//_ = b.createDefaultCalendar(context.Background())
	return b, nil
}

func (b *backend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	upPath, err := b.CurrentUserPrincipal(ctx)
	if err != nil {
		return "", err
	}

	return path.Join(upPath, b.prefix) + "/", nil
}

func (b *backend) CreateCalendar(ctx context.Context, calendar *caldav.Calendar) error {
	homeSetPath, err := b.CalendarHomeSetPath(ctx)
	if err != nil {
		return err
	}
	if err := b.repo.CreateFolder(ctx, homeSetPath, calendar); err != nil {
		return err
	}
	return nil
}

func (b *backend) createDefaultCalendar(ctx context.Context) error {
	if err := b.CreateCalendar(ctx, &caldav.Calendar{
		Name:                  "Private",
		Description:           "Test",
		MaxResourceSize:       4096,
		SupportedComponentSet: []string{"VEVENT", "VTODO"},
	}); err != nil {
		return err
	}
	return nil
}

func (b *backend) ListCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	cals, err := b.repo.FindFolders(ctx)
	if err != nil {
		return nil, err
	}

	for i, cal := range cals {
		homeSetPath, err := b.CalendarHomeSetPath(ctx)
		if err != nil {
			return make([]caldav.Calendar, 0), err
		}
		cals[i].Path = path.Join(homeSetPath, cal.Path) + "/"
	}
	return cals, nil
}

func (b *backend) GetCalendar(ctx context.Context, urlPath string) (*caldav.Calendar, error) {
	cals, err := b.repo.FindFolders(ctx)
	if err != nil {
		return nil, err
	}

	for _, cal := range cals {
		homeSetPath, err := b.CalendarHomeSetPath(ctx)
		if err != nil {
			return nil, err
		}
		if path.Join(homeSetPath, cal.Path)+"/" == urlPath {
			return &cal, nil
		}
	}
	return nil, fmt.Errorf("calendar for path: %s not found", urlPath)
}

func (b *backend) GetCalendarObject(
	ctx context.Context,
	objPath string,
	req *caldav.CalendarCompRequest,
) (*caldav.CalendarObject, error) {
	var propFilter []string
	if req != nil && !req.AllProps {
		propFilter = req.Props
	}

	uid := strings.TrimSuffix(path.Base(objPath), ".ics")

	cal, err := b.repo.GetCalendar(ctx, uid, propFilter)
	if err != nil {
		return nil, err
	}
	obj, err := b.repo.GetFileInfo(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("object for path: %s not found", objPath)
	}

	var buf bytes.Buffer
	f := bufio.NewWriter(&buf)

	enc := ical.NewEncoder(f)
	err = enc.Encode(cal)
	if err != nil {
		return nil, err
	}

	err = f.Flush()
	if err != nil {
		return nil, err
	}

	length := int64(buf.Len())

	obj.Path = objPath
	obj.Data = cal

	if obj.ContentLength != length {
		obj.ContentLength = length
	}

	return obj, nil
}

func (b *backend) ListCalendarObjects(
	ctx context.Context,
	urlPath string,
	req *caldav.CalendarCompRequest,
) ([]caldav.CalendarObject, error) {
	var propFilter []string
	if req != nil && !req.AllProps {
		propFilter = req.Props
	}

	homeSetPath, _ := b.CalendarHomeSetPath(ctx)
	folderID, err := strconv.Atoi(path.Base(urlPath))
	if err != nil {
		return nil, fmt.Errorf("invalid folder_id: %s", urlPath)
	}
	objs, err := b.repo.FindObjects(ctx, folderID, propFilter)
	if err != nil {
		return nil, err
	}

	for i, obj := range objs {
		uid := strings.TrimSuffix(path.Base(obj.Path), ".ics")

		cal, err := b.repo.GetCalendar(ctx, uid, propFilter)
		if err != nil {
			return nil, err
		}

		objs[i].Data = cal
		objs[i].Path = path.Join(homeSetPath, strconv.Itoa(folderID), uid+".ics")
	}
	return objs, nil
}

func (b *backend) QueryCalendarObjects(
	ctx context.Context,
	urlPath string,
	query *caldav.CalendarQuery,
) ([]caldav.CalendarObject, error) {
	var propFilter []string
	if query != nil && !query.CompRequest.AllProps {
		propFilter = query.CompRequest.Props
	}

	homeSetPath, _ := b.CalendarHomeSetPath(ctx)
	folderID, err := strconv.Atoi(path.Base(urlPath))
	if err != nil {
		return nil, fmt.Errorf("invalid folder_id: %s", urlPath)
	}
	objs, err := b.repo.FindObjects(ctx, folderID, propFilter)
	if err != nil {
		return nil, err
	}

	for i, obj := range objs {
		uid := strings.TrimSuffix(path.Base(obj.Path), ".ics")

		cal, err := b.repo.GetCalendar(ctx, uid, propFilter)
		if err != nil {
			return nil, err
		}

		objs[i].Data = cal
		objs[i].Path = path.Join(homeSetPath, strconv.Itoa(folderID), uid+".ics")
	}
	return caldav.Filter(query, objs)
}

func (b *backend) PutCalendarObject(
	ctx context.Context,
	objPath string,
	calendar *ical.Calendar,
	opts *caldav.PutCalendarObjectOptions,
) (*caldav.CalendarObject, error) {
	eventType, uid, err := caldav.ValidateCalendarObject(calendar)
	if err != nil {
		return nil, caldav.NewPreconditionError(caldav.PreconditionValidCalendarObjectResource)
	}
	// Object always get saved as <UID>.ics
	dirname, _ := path.Split(objPath)
	objPath = path.Join(dirname, uid+".ics")

	var tzIndex int
	shouldRemoveTimezone := false

	for i, child := range calendar.Component.Children {
		if child.Name == ical.CompTimezone {
			tzIndex = i
			shouldRemoveTimezone = true
		}
		if child.Name == eventType {
			for _, v := range child.Props {
				if v[0].ValueType() == ical.ValueDateTime {
					oldTime, _ := v[0].DateTime(time.UTC)
					v[0].SetDateTime(oldTime.UTC())
					delete(v[0].Params, ical.ParamTimezoneID)
				}
			}
		}
	}

	if shouldRemoveTimezone {
		calendar.Component.Children = append(
			calendar.Component.Children[:tzIndex],
			calendar.Component.Children[tzIndex+1:]...,
		)
	}

	var buf bytes.Buffer
	f := bufio.NewWriter(&buf)

	enc := ical.NewEncoder(f)
	err = enc.Encode(calendar)
	if err != nil {
		return nil, err
	}

	err = f.Flush()
	if err != nil {
		return nil, err
	}

	eTag, err := etag.FromData(buf.Bytes())
	if err != nil {
		return nil, err
	}

	obj := &caldav.CalendarObject{
		Path:          objPath,
		ContentLength: int64(buf.Len()),
		Data:          calendar,
		ETag:          eTag,
		ModTime:       time.Now().UTC(),
	}
	return b.repo.PutObject(ctx, uid, eventType, obj, opts)
}

func (b *backend) DeleteCalendarObject(ctx context.Context, path string) error {
	//delete(b.objectMap, path)
	return nil
}
