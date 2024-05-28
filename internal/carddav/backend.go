package carddav

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/carddav"
)

type Backend struct {
	webdav.UserPrincipalBackend
	Prefix string
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

func New(upBackend webdav.UserPrincipalBackend, prefix string, repository Repository) (*Backend, error) {
	return &Backend{
		UserPrincipalBackend: upBackend,
		Prefix:               prefix,
		repo:                 repository,
	}, nil
}

func (b *Backend) AddressBookHomeSetPath(ctx context.Context) (string, error) {
	upPath, err := b.CurrentUserPrincipal(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(upPath, b.Prefix) + "/", nil
}

func (b *Backend) ListAddressBooks(ctx context.Context) ([]carddav.AddressBook, error) {
	datatype := make([]carddav.AddressDataType, 0)
	datatype = append(datatype,
		carddav.AddressDataType{
			ContentType: "text/vcard",
			Version:     "3.0",
		},
		carddav.AddressDataType{
			ContentType: "text/vcard",
			Version:     "4.0",
		},
		carddav.AddressDataType{
			ContentType: "application/vcard",
			Version:     "4.0",
		},
	)
	//p := ctx.Value(addressBookPathKey).(string)
	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return nil, err
	}

	//addrbook := carddav.AddressBook{
	//	Path:            path.Join(homeSetPath, "-/"),
	//	Name:            "My contacts",
	//	Description:     "Default address book",
	//	MaxResourceSize: 10 * 1024,
	//}

	return []carddav.AddressBook{
		{
			Path:                 homeSetPath + "1/",
			Name:                 "My contacts1",
			Description:          "Default address book",
			MaxResourceSize:      10 * 1024,
			SupportedAddressData: datatype,
		},

		{
			Path:                 homeSetPath + "2/",
			Name:                 "My contacts2",
			Description:          "Default address book",
			MaxResourceSize:      10 * 1024,
			SupportedAddressData: datatype,
		},
	}, nil
}

func (b *Backend) GetAddressBook(ctx context.Context, urlPath string) (*carddav.AddressBook, error) {
	datatype := make([]carddav.AddressDataType, 0)
	datatype = append(datatype,
		carddav.AddressDataType{
			ContentType: "text/vcard",
			Version:     "3.0",
		},
		carddav.AddressDataType{
			ContentType: "text/vcard",
			Version:     "4.0",
		},
		carddav.AddressDataType{
			ContentType: "application/vcard",
			Version:     "4.0",
		},
	)
	return &carddav.AddressBook{
		Path:                 urlPath,
		Name:                 "My contacts",
		Description:          "Default address book",
		MaxResourceSize:      10 * 1024,
		SupportedAddressData: datatype,
	}, nil
	//panic("TODO: implement GetAddressBook")
}

func (b *Backend) CreateAddressBook(ctx context.Context, addressBook *carddav.AddressBook) error {
	panic("TODO: implement CreateAddressBook")
}

func (b *Backend) DeleteAddressBook(ctx context.Context, path string) error {
	panic("TODO: implement DeleteAddressBook")
}

func (b *Backend) GetAddressObject(ctx context.Context, path string, req *carddav.AddressDataRequest) (*carddav.AddressObject, error) {
	if path == "/admin/contacts/default/" {
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

func (b *Backend) ListAddressObjects(ctx context.Context, path string, req *carddav.AddressDataRequest) ([]carddav.AddressObject, error) {
	alice, err := b.GetAddressObject(ctx, "/admin/contacts/default/", req)
	if err != nil {
		return nil, err
	}
	return []carddav.AddressObject{*alice}, nil
	//panic("TODO: implement")
}

func (b *Backend) QueryAddressObjects(ctx context.Context, path string, query *carddav.AddressBookQuery) ([]carddav.AddressObject, error) {
	panic("TODO: implement QueryAddressObjects")
}

func (b *Backend) PutAddressObject(ctx context.Context, path string, card vcard.Card, opts *carddav.PutAddressObjectOptions) (*carddav.AddressObject, error) {
	panic("TODO: implement PutAddressObject")
}

func (b *Backend) DeleteAddressObject(ctx context.Context, path string) error {
	panic("TODO: implement DeleteAddressObject")
}
