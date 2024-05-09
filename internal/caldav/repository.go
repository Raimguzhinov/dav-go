package caldav

import (
	"context"

	"github.com/emersion/go-webdav/caldav"
)

type Repository interface {
	CreateFolder(ctx context.Context, calendar *caldav.Calendar) error
	FindFolders(ctx context.Context) ([]caldav.Calendar, error)
}
