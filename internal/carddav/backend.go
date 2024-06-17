package carddav

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/carddav"
	"github.com/emersion/go-vcard"
	"github.com/google/uuid"
)

type backend struct {
	webdav.UserPrincipalBackend
	prefix string
	repo   Repository
}

var (
	aliceData = `BEGIN:VCARD
VERSION:4.0
UID:urn:uuid:4fbe8971-0bc3-424c-9c26-36c3e1eff6b1
FN;PID=1.1:Alice Gopher
N:Gopher;Alice;;;
EMAIL;PID=1.1:alice@example.com
CLIENTPIDMAP:1;urn:uuid:53e374d9-337e-4727-8803-a1e9c14e0551
END:VCARD`
	//alicePath = "urn:uuid:4fbe8971-0bc3-424c-9c26-36c3e1eff6b1.vcf"
	//
	//currentUserPrincipalKey = contextKey("test:currentUserPrincipal")
	//homeSetPathKey          = contextKey("test:homeSetPath")
	//addressBookPathKey      = contextKey("test:addressBookPath")
)

func New(upBackend webdav.UserPrincipalBackend, prefix string, repository Repository) (*backend, error) {
	return &backend{
		UserPrincipalBackend: upBackend,
		prefix:               prefix,
		repo:                 repository,
	}, nil
}

func (b *backend) AddressBookHomeSetPath(ctx context.Context) (string, error) {
	upPath, err := b.CurrentUserPrincipal(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(upPath, b.prefix) + "/", nil
}

func (b *backend) CreateDefaultAddressBook(ctx context.Context) (*carddav.AddressBook, error) {
	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return nil, err
	}

	uid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	ab := carddav.AddressBook{
		Path:        path.Join(homeSetPath, uid.String(), "/"),
		Name:        "Contacts",
		Description: "Default addressbook",
	}
	err = b.repo.CreateFolder(ctx, homeSetPath, &ab)
	if err != nil {
		return nil, err
	}
	return &ab, nil
}

func (b *backend) ListAddressBooks(ctx context.Context) ([]carddav.AddressBook, error) {
	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return nil, err
	}

	addressbooks, err := b.repo.FindFolders(ctx, homeSetPath)
	if err != nil {
		return nil, err
	}

	if len(addressbooks) == 0 {
		defaultAB, err := b.CreateDefaultAddressBook(ctx)
		if err != nil {
			return nil, err
		}
		return []carddav.AddressBook{
			*defaultAB,
		}, nil
	}

	for i := range addressbooks {
		addressbooks[i].Path = path.Join(homeSetPath, addressbooks[i].Path) + "/"
	}

	return addressbooks, nil
}

func (b *backend) GetAddressBook(ctx context.Context, urlPath string) (*carddav.AddressBook, error) {
	urlPath = strings.TrimSuffix(urlPath, "/")

	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return nil, err
	}

	addressbooks, err := b.repo.FindFolders(ctx, homeSetPath)
	if err != nil {
		return nil, err
	}

	for _, addressbook := range addressbooks {
		if path.Join(homeSetPath, addressbook.Path) == urlPath {
			addressbook.Path = path.Join(homeSetPath, addressbook.Path) + "/"
			return &addressbook, nil
		}
	}
	return nil, fmt.Errorf("addressbook for path: %s not found", urlPath)
}

func (b *backend) CreateAddressBook(ctx context.Context, addressBook *carddav.AddressBook) error {
	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return err
	}

	//addressBook.Path = strings.TrimSuffix(strings.TrimPrefix(addressBook.Path, homeSetPath), "/")
	err = b.repo.CreateFolder(ctx, homeSetPath, addressBook)
	if err != nil {
		return err
	}
	return nil
}

func (b *backend) DeleteAddressBook(ctx context.Context, path string) error {
	addressbook, err := b.GetAddressBook(ctx, path)
	if err != nil {
		return err
	}
	err = b.repo.DeleteFolder(ctx, addressbook)
	if err != nil {
		return err
	}
	return nil
}

func (b *backend) GetAddressObject(ctx context.Context, path string, req *carddav.AddressDataRequest) (*carddav.AddressObject, error) {
	// TODO
	if path == "/admin/contacts/1/" {
		card, err := vcard.NewDecoder(strings.NewReader(aliceData)).Decode()
		if err != nil {
			return nil, err
		}
		return &carddav.AddressObject{
			Path: path,
			Card: card,
		}, nil
	} else {
		return nil, webdav.NewHTTPError(404, fmt.Errorf("Not found"))
	}
	//panic("TODO: implement GetAddressObject")
}

func (b *backend) ListAddressObjects(ctx context.Context, path string, req *carddav.AddressDataRequest) ([]carddav.AddressObject, error) {

	alice, err := b.GetAddressObject(ctx, "/admin/contacts/1/", req)
	if err != nil {
		return nil, err
	}
	return []carddav.AddressObject{*alice}, nil

}

func (b *backend) QueryAddressObjects(ctx context.Context, path string, query *carddav.AddressBookQuery) ([]carddav.AddressObject, error) {

	return nil, nil
}

func (b *backend) PutAddressObject(ctx context.Context, path string, card vcard.Card, opts *carddav.PutAddressObjectOptions) (*carddav.AddressObject, error) {

	ao := carddav.AddressObject{
		Path: path,
		Card: card,
	}
	_, err := b.repo.PutAddressObject(ctx, &ao, opts)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *backend) DeleteAddressObject(ctx context.Context, path string) error {

	return nil
}

func (b *backend) GetCurrentUserAddressBookPrivilege(ctx context.Context, ab *carddav.AddressBook) ([]string, error) {
	return []string{"all", "read", "write", "write-properties", "write-content", "unlock", "bind", "unbind", "write-acl", "read-acl", "read-current-user-privilege-set"}, nil
}
