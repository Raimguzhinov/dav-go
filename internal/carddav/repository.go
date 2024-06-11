package carddav

import (
	"context"

	"github.com/ceres919/go-webdav/carddav"
)

type Repository interface {
	CreateFolder(ctx context.Context, homeSetPath string, addressbook *carddav.AddressBook) error
	FindFolders(ctx context.Context) ([]carddav.AddressBook, error)
	DeleteFolder(ctx context.Context, addressbook *carddav.AddressBook) error
	PutObject(ctx context.Context, uid string, object *carddav.AddressObject, opts *carddav.PutAddressObjectOptions) (string, error)
	CreateContact(ctx context.Context, addressbook *carddav.AddressBook) error
	FindContacts(ctx context.Context) ([]carddav.AddressObject, error)
}
