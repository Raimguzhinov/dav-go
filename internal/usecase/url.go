package usecase

import (
	"fmt"
	"net/url"

	"github.com/Raimguhinov/dav-go/internal/usecase/repo"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

type UseCaseUrl struct {
	storageURL    string
	caldavPrefix  string
	carddavPrefix string
	upBackend     webdav.UserPrincipalBackend
}

func NewURL(
	storageURL, caldavPrefix, carddavPrefix string,
	upBackend webdav.UserPrincipalBackend,
) *UseCaseUrl {
	return &UseCaseUrl{
		storageURL:    storageURL,
		caldavPrefix:  caldavPrefix,
		carddavPrefix: carddavPrefix,
		upBackend:     upBackend,
	}
}

func NewFromURL(
	useCaseUrl *UseCaseUrl,
	provider any,
	logger *logger.Logger,
) (caldav.Backend, carddav.Backend, error) {
	u, err := url.Parse(useCaseUrl.storageURL)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing storage URL: %s", err.Error())
	}

	switch u.Scheme {
	case "postgres":
		pg, ok := provider.(*postgres.Postgres)
		if !ok {
			return nil, nil, fmt.Errorf("postgres provider not supported")
		}
		return repo.NewBackends(
			useCaseUrl.upBackend,
			useCaseUrl.caldavPrefix,
			useCaseUrl.carddavPrefix,
			pg,
			logger,
		)
	default:
		return nil, nil, fmt.Errorf("no storage provider found for %s:// URL", u.Scheme)
	}
}
