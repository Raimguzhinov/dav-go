package carddav_db

import (
	"context"

	backend "github.com/Raimguhinov/dav-go/internal/carddav"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/ceres919/go-webdav/carddav"
)

type repository struct {
	client *postgres.Postgres
	logger *logger.Logger
}

func NewRepository(client *postgres.Postgres, logger *logger.Logger) backend.Repository {
	return &repository{
		client: client,
		logger: logger,
	}
}

func (r *repository) CreateFolder(ctx context.Context, addressbook *carddav.AddressBook) error {
	panic("TODO")
}

func (r *repository) FindFolders(ctx context.Context) ([]carddav.AddressBook, error) {
	panic("TODO")
}

func (r *repository) PutObject(ctx context.Context, uid string, object *carddav.AddressObject, opts *carddav.PutAddressObjectOptions) (string, error) {
	panic("TODO")
}

func (r *repository) CreateContact(ctx context.Context, addressbook *carddav.AddressBook) error {
	panic("TODO")
}
