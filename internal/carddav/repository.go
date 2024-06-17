package carddav

import (
	"context"

	"github.com/ceres919/go-webdav/carddav"
)

type Repository interface {
	CreateFolder(ctx context.Context, homeSetPath string, addressbook *carddav.AddressBook) error
	FindFolders(ctx context.Context, homeSetPath string) ([]carddav.AddressBook, error)
	DeleteFolder(ctx context.Context, addressbook *carddav.AddressBook) error
	PutAddressObject(ctx context.Context, object *carddav.AddressObject, opts *carddav.PutAddressObjectOptions) (string, error)
	FindAddressObjects(ctx context.Context) ([]carddav.AddressObject, error)
	DeleteAddressObject(ctx context.Context, path string) error
	GetFolderAccess(ctx context.Context, addressbook *carddav.AddressBook) ([]string, error)
}
