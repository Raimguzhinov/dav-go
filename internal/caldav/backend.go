package caldav

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

type Backend struct {
	webdav.UserPrincipalBackend
	Prefix string
	repo   Repository
}

func New(upBackend webdav.UserPrincipalBackend, prefix string, repository Repository) (*Backend, error) {
	b := &Backend{
		UserPrincipalBackend: upBackend,
		Prefix:               prefix,
		repo:                 repository,
	}

	//if err := b.repo.CreateFolder(
	//	context.Background(),
	//	&caldav.Calendar{Name: "Calendar", SupportedComponentSet: []string{"VEVENT", "VTODO"}},
	//); err != nil {
	//	return nil, err
	//}
	return b, nil
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
		cals[i].Path = path.Join(homeSetPath, cal.Path)
	}
	return cals, nil
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
	return nil, fmt.Errorf("calendar for path: %s not found", path)
}

func (b *Backend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	upPath, err := b.CurrentUserPrincipal(ctx)
	if err != nil {
		return "", err
	}

	return path.Join(upPath, b.Prefix) + "/", nil
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

func (b *Backend) PutCalendarObject(ctx context.Context, objPath string, calendar *ical.Calendar, opts *caldav.PutCalendarObjectOptions) (string, error) {
	eventType, uid, err := caldav.ValidateCalendarObject(calendar)
	if err != nil {
		return "", caldav.NewPreconditionError(caldav.PreconditionValidCalendarObjectResource)
	}
	// Object always get saved as <UID>.ics
	dirname, _ := path.Split(objPath)
	objPath = path.Join(dirname, uid+".ics")

	var buf bytes.Buffer
	f := bufio.NewWriter(&buf)

	enc := ical.NewEncoder(f)
	err = enc.Encode(calendar)
	if err != nil {
		return "", err
	}

	err = f.Flush()
	if err != nil {
		return "", err
	}

	size := int64(buf.Len())
	etag, err := EtagForData(buf.Bytes())
	if err != nil {
		return "", err
	}

	object := caldav.CalendarObject{
		Path:          objPath,
		ContentLength: size,
		Data:          calendar,
		ETag:          etag,
		ModTime:       time.Now().UTC(),
	}
	return b.repo.PutObject(ctx, uid, eventType, &object, opts)
}

func EtagForData(data []byte) (string, error) {
	h := sha1.New()
	if _, err := io.Copy(h, bytes.NewReader(data)); err != nil {
		return "", err
	}
	csum := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(csum[:]), nil
}

func (b *Backend) ListCalendarObjects(ctx context.Context, path string, req *caldav.CalendarCompRequest) ([]caldav.CalendarObject, error) {
	return nil, nil
}

func (b *Backend) QueryCalendarObjects(ctx context.Context, query *caldav.CalendarQuery) ([]caldav.CalendarObject, error) {
	return nil, nil
}
