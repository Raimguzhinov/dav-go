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

	"github.com/Raimguhinov/dav-go/internal/delivery/grpc"
	"github.com/Raimguhinov/dav-go/internal/usecase/etag"
	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/caldav"
	"github.com/emersion/go-ical"
	"github.com/google/uuid"
)

type backend struct {
	webdav.UserPrincipalBackend
	prefix string
	repo   Repository
	s      grpc.CalendarServer
}

func New(
	upBackend webdav.UserPrincipalBackend,
	prefix string,
	repository Repository,
) (caldav.Backend, error) {
	s := &backend{
		UserPrincipalBackend: upBackend,
		prefix:               prefix,
		repo:                 repository,
	}
	//_ = s.createDefaultCalendar(context.Background())
	return s, nil
}

func (s *backend) CalendarHomeSetPath(ctx context.Context) (string, error) {
	upPath, err := s.CurrentUserPrincipal(ctx)
	if err != nil {
		return "", err
	}

	return path.Join(upPath, s.prefix) + "/", nil
}

func (s *backend) CreateCalendar(ctx context.Context, calendar *caldav.Calendar) error {
	homeSetPath, err := s.CalendarHomeSetPath(ctx)
	if err != nil {
		return err
	}
	if calendar.MaxResourceSize == 0 || calendar.SupportedComponentSet == nil {
		return s.createDefaultCalendar(ctx, calendar.Name)
	}
	if err := s.repo.CreateCalendar(ctx, homeSetPath, calendar); err != nil {
		return err
	}
	return nil
}

func (s *backend) createDefaultCalendar(ctx context.Context, name string) error {
	if err := s.CreateCalendar(ctx, &caldav.Calendar{
		Name:                  name,
		Description:           "Protei Calendar",
		MaxResourceSize:       4096,
		SupportedComponentSet: []string{"VEVENT", "VTODO", "VJOURNAL"},
	}); err != nil {
		return err
	}
	return nil
}

func (s *backend) ListCalendars(ctx context.Context) ([]caldav.Calendar, error) {
	cals, err := s.repo.FindCalendars(ctx)
	if err != nil {
		return nil, err
	}

	for i, cal := range cals {
		homeSetPath, err := s.CalendarHomeSetPath(ctx)
		if err != nil {
			return make([]caldav.Calendar, 0), err
		}
		cals[i].Path = path.Join(homeSetPath, cal.Path) + "/"
	}
	return cals, nil
}

func (s *backend) GetCalendar(ctx context.Context, urlPath string) (*caldav.Calendar, error) {
	cals, err := s.repo.FindCalendars(ctx)
	if err != nil {
		return nil, err
	}

	for _, cal := range cals {
		homeSetPath, err := s.CalendarHomeSetPath(ctx)
		if err != nil {
			return nil, err
		}
		if path.Join(homeSetPath, cal.Path)+"/" == urlPath {
			return &cal, nil
		}
	}
	return nil, fmt.Errorf("calendar for path: %s not found", urlPath)
}

func (s *backend) GetCalendarObject(
	ctx context.Context,
	objPath string,
	req *caldav.CalendarCompRequest,
) (*caldav.CalendarObject, error) {
	var propFilter []string
	if req != nil && !req.AllProps {
		propFilter = req.Props
	}
	uid := strings.TrimSuffix(path.Base(objPath), ".ics")
	if err := uuid.Validate(uid); err != nil {
		return nil, fmt.Errorf("object for path: %s not found", objPath)
	}

	cal, err := s.repo.GetCalendar(ctx, uid, propFilter)
	if err != nil {
		return nil, err
	}
	obj, err := s.repo.GetCalendarObjectInfo(ctx, uid)
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

func (s *backend) ListCalendarObjects(
	ctx context.Context,
	urlPath string,
	req *caldav.CalendarCompRequest,
) ([]caldav.CalendarObject, error) {
	var propFilter []string
	if req != nil && !req.AllProps {
		propFilter = req.Props
	}

	homeSetPath, _ := s.CalendarHomeSetPath(ctx)
	folderID, err := strconv.Atoi(path.Base(urlPath))
	if err != nil {
		return nil, fmt.Errorf("invalid folder_id: %s", urlPath)
	}
	objs, err := s.repo.FindCalendarObjects(ctx, folderID, propFilter)
	if err != nil {
		return nil, err
	}

	for i, obj := range objs {
		uid := strings.TrimSuffix(path.Base(obj.Path), ".ics")

		cal, err := s.repo.GetCalendar(ctx, uid, propFilter)
		if err != nil {
			return nil, err
		}

		objs[i].Data = cal
		objs[i].Path = path.Join(homeSetPath, strconv.Itoa(folderID), uid+".ics")
	}
	return objs, nil
}

func (s *backend) QueryCalendarObjects(
	ctx context.Context,
	urlPath string,
	query *caldav.CalendarQuery,
) ([]caldav.CalendarObject, error) {
	var propFilter []string
	if query != nil && !query.CompRequest.AllProps {
		propFilter = query.CompRequest.Props
	}

	homeSetPath, _ := s.CalendarHomeSetPath(ctx)
	folderID, err := strconv.Atoi(path.Base(urlPath))
	if err != nil {
		return nil, fmt.Errorf("invalid folder_id: %s", urlPath)
	}
	objs, err := s.repo.FindCalendarObjects(ctx, folderID, propFilter)
	if err != nil {
		return nil, err
	}

	for i, obj := range objs {
		uid := strings.TrimSuffix(path.Base(obj.Path), ".ics")

		cal, err := s.repo.GetCalendar(ctx, uid, propFilter)
		if err != nil {
			return nil, err
		}

		objs[i].Data = cal
		objs[i].Path = path.Join(homeSetPath, strconv.Itoa(folderID), uid+".ics")
	}
	return caldav.Filter(query, objs)
}

func (s *backend) PutCalendarObject(
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
	return s.repo.UpgradeCalendarObject(ctx, uid, eventType, obj, opts)
}

func (s *backend) DeleteCalendarObject(ctx context.Context, path string) error {
	//delete(s.objectMap, path)
	return nil
}

func (s *backend) GetPrivileges(ctx context.Context) []string {
	return []string{"all", "read", "write", "write-properties", "write-content", "unlock", "bind", "unbind", "write-acl", "read-acl", "read-current-user-privilege-set"}
}

func (s *backend) GetCalendarPrivileges(ctx context.Context, cal *caldav.Calendar) []string {
	return []string{"all", "read", "write", "write-properties", "write-content", "unlock", "bind", "unbind", "write-acl", "read-acl", "read-current-user-privilege-set"}
}
