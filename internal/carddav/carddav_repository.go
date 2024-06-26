package carddav

import (
	"context"

	"github.com/ceres919/go-webdav/carddav"
)

type RepositoryCarddav interface {
	CreateFolder(ctx context.Context, homeSetPath string, addressbook *carddav.AddressBook) error
	FindFolders(ctx context.Context, homeSetPath string) ([]carddav.AddressBook, error)
	DeleteFolder(ctx context.Context, addressbook *carddav.AddressBook) error
	PutAddressObject(ctx context.Context, homeSetPath string, object *carddav.AddressObject, opts *carddav.PutAddressObjectOptions) error
	FindAddressObjects(ctx context.Context, homeSetPath, abUID string) ([]carddav.AddressObject, error)
	DeleteAddressObject(ctx context.Context, urlPath string) error
	GetFolderAccess(ctx context.Context, addressbook *carddav.AddressBook) ([]string, error)
}
