package carddav

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/ceres919/go-webdav"
	"github.com/ceres919/go-webdav/carddav"
	"github.com/emersion/go-vcard"
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

func (b *backend) ListAddressBooks(ctx context.Context) ([]carddav.AddressBook, error) {
	addressbooks, err := b.repo.FindFolders(ctx)
	if err != nil {
		return nil, err
	}
	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return nil, err
	}

	for _, addressbook := range addressbooks {
		addressbook.Path = path.Join(homeSetPath, addressbook.Path) + "/"
	}
	return addressbooks, nil
}

func (b *backend) GetAddressBook(ctx context.Context, urlPath string) (*carddav.AddressBook, error) {
	addressbooks, err := b.repo.FindFolders(ctx)
	if err != nil {
		return nil, err
	}
	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return nil, err
	}

	for _, addressbook := range addressbooks {
		if path.Join(homeSetPath, addressbook.Path)+"/" == urlPath {
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
	// TODO
	alice, err := b.GetAddressObject(ctx, "/admin/contacts/default/", req)
	if err != nil {
		return nil, err
	}
	return []carddav.AddressObject{*alice}, nil
	//panic("TODO: implement")
}

func (b *backend) QueryAddressObjects(ctx context.Context, path string, query *carddav.AddressBookQuery) ([]carddav.AddressObject, error) {
	// TODO
	panic("TODO: implement QueryAddressObjects")
}

func (b *backend) PutAddressObject(ctx context.Context, path string, card vcard.Card, opts *carddav.PutAddressObjectOptions) (*carddav.AddressObject, error) {
	// TODO

	panic("TODO: implement PutAddressObject")
}

func (b *backend) DeleteAddressObject(ctx context.Context, path string) error {
	// TODO
	panic("TODO: implement DeleteAddressObject")
}
