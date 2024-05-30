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

type Backend struct {
	webdav.UserPrincipalBackend
	Prefix string
	repo   Repository
}

func New(
	upBackend webdav.UserPrincipalBackend,
	prefix string,
	repository Repository,
) (*Backend, error) {
	b := &Backend{
		UserPrincipalBackend: upBackend,
		Prefix:               prefix,
		repo:                 repository,
	}
	//_ = b.createDefaultCalendar(context.Background())
	return b, nil
}

func (b *Backend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	upPath, err := b.CurrentUserPrincipal(ctx)
	if err != nil {
		return "", err
	}

	return path.Join(upPath, b.Prefix) + "/", nil
}

func (b *Backend) CreateCalendar(ctx context.Context, calendar *caldav.Calendar) error {
	homeSetPath, err := b.CalendarHomeSetPath(ctx)
	if err != nil {
		return err
	}
	if err := b.repo.CreateFolder(ctx, homeSetPath, calendar); err != nil {
		return err
	}
	return nil
}

func (b *Backend) createDefaultCalendar(ctx context.Context) error {
	if err := b.CreateCalendar(ctx, &caldav.Calendar{
		Name:                  "Alien",
		Description:           "Test",
		MaxResourceSize:       4096,
		SupportedComponentSet: []string{"VEVENT", "VTODO"},
	}); err != nil {
		return err
	}
	return nil
}

func (b *Backend) ListCalendars(ctx context.Context) ([]caldav.Calendar, error) {
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

func (b *Backend) GetCalendar(ctx context.Context, urlPath string) (*caldav.Calendar, error) {
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

func (b *Backend) GetCalendarObject(
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

	obj.Path = objPath
	obj.Data = cal

	return obj, err
}

func (b *Backend) ListCalendarObjects(
	ctx context.Context,
	urlPath string,
	req *caldav.CalendarCompRequest,
) ([]caldav.CalendarObject, error) {
	var propFilter []string
	if req != nil && !req.AllProps {
		propFilter = req.Props
	}
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
	}
	return objs, nil
}

func (b *Backend) QueryCalendarObjects(
	ctx context.Context,
	urlPath string,
	query *caldav.CalendarQuery,
) ([]caldav.CalendarObject, error) {
	var propFilter []string
	if query != nil && !query.CompRequest.AllProps {
		propFilter = query.CompRequest.Props
	}

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
	}
	return caldav.Filter(query, objs)
}

func (b *Backend) PutCalendarObject(
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

	size := int64(buf.Len())
	eTag, err := etag.FromData(buf.Bytes())
	if err != nil {
		return nil, err
	}

	obj := &caldav.CalendarObject{
		Path:          objPath,
		ContentLength: size,
		Data:          calendar,
		ETag:          eTag,
		ModTime:       time.Now().UTC(),
	}
	return b.repo.PutObject(ctx, uid, eventType, obj, opts)
}

func (b *Backend) DeleteCalendarObject(ctx context.Context, path string) error {
	//delete(b.objectMap, path)
	return nil
}
