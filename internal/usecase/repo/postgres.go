package repo

import (
	caldavBackend "github.com/Raimguzhinov/dav-go/internal/caldav"
	caldavDB "github.com/Raimguzhinov/dav-go/internal/caldav/db"
	carddavBackend "github.com/Raimguzhinov/dav-go/internal/carddav"
	"github.com/Raimguzhinov/dav-go/internal/carddav/db"
	"github.com/Raimguzhinov/dav-go/pkg/logger"
	"github.com/Raimguzhinov/dav-go/pkg/postgres"
	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/caldav"
	"github.com/ceres919/go-webdav/carddav"
)

func NewBackends(
	upBackend webdav.UserPrincipalBackend,
	caldavPrefix, carddavPrefix string,
	pg *postgres.Postgres,
	logger *logger.Logger,
) (caldav.Backend, carddav.Backend, error) {
	calBackend, err := caldavBackend.New(
		upBackend,
		caldavPrefix,
		caldavDB.NewRepository(pg, logger),
	)
	if err != nil {
		return nil, nil, err
	}
	cardBackend, err := carddavBackend.New(
		upBackend,
		carddavPrefix,
		db.NewRepository(pg, logger),
	)
	if err != nil {
		return nil, nil, err
	}
	return calBackend, cardBackend, err
}
