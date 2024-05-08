package storage

import (
	"fmt"
	"net/url"

	"github.com/Raimguhinov/dav-go/internal/storage/postgres"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

func NewFromURL(storageURL, caldavPrefix, carddavPrefix string, userPrincipalBackend webdav.UserPrincipalBackend) (caldav.Backend, carddav.Backend, error) {
	u, err := url.Parse(storageURL)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing storage URL: %s", err.Error())
	}

	switch u.Scheme {
	case "postgres":
		return postgres.New(caldavPrefix, carddavPrefix, userPrincipalBackend)
	default:
		return nil, nil, fmt.Errorf("no storage provider found for %s:// URL", u.Scheme)
	}
}
