package repo

import (
	caldavBackend "github.com/Raimguhinov/dav-go/internal/caldav"
	"github.com/Raimguhinov/dav-go/internal/caldav/db"
	carddavBackend "github.com/Raimguhinov/dav-go/internal/carddav"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

func NewBackends(pg *postgres.Postgres, logger *logger.Logger) (caldav.Backend, carddav.Backend, error) {
	calBackend, err := caldavBackend.New(caldav_db.NewRepository(pg, logger))
	if err != nil {
		return nil, nil, err
	}
	cardBackend, err := carddavBackend.New()
	if err != nil {
		return nil, nil, err
	}
	return calBackend, cardBackend, err
}
