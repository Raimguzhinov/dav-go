package postgres

import (
	"context"

	caldavBackend "github.com/Raimguhinov/dav-go/internal/caldav"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

func New(caldavPrefix, carddavPrefix string, userPrincipalBackend webdav.UserPrincipalBackend) (caldav.Backend, carddav.Backend, error) {
	_ = caldavPrefix
	_ = carddavPrefix
	name, err := userPrincipalBackend.CurrentUserPrincipal(context.TODO())
	if err != nil {
		return nil, nil, err
	}
	_ = name
	return caldavBackend.New(), nil, nil
}
