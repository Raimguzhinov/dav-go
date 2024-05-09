package carddav

import (
	"context"
	"fmt"
	"strings"

	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/carddav"
)

type Backend struct {
}

func New() (*Backend, error) {
	return &Backend{}, nil
}

type contextKey string

var (
	aliceData = `BEGIN:VCARD
VERSION:4.0
UID:urn:uuid:4fbe8971-0bc3-424c-9c26-36c3e1eff6b1
FN;PID=1.1:Alice Gopher
N:Gopher;Alice;;;
EMAIL;PID=1.1:alice@example.com
CLIENTPIDMAP:1;urn:uuid:53e374d9-337e-4727-8803-a1e9c14e0551
END:VCARD`
	alicePath = "urn:uuid:4fbe8971-0bc3-424c-9c26-36c3e1eff6b1.vcf"

	currentUserPrincipalKey = contextKey("test:currentUserPrincipal")
	homeSetPathKey          = contextKey("test:homeSetPath")
	addressBookPathKey      = contextKey("test:addressBookPath")
)

func (*Backend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	r := ctx.Value(currentUserPrincipalKey).(string)
	return r, nil
}

func (*Backend) AddressbookHomeSetPath(ctx context.Context) (string, error) {
	r := ctx.Value(homeSetPathKey).(string)
	return r, nil
}

func (*Backend) AddressBook(ctx context.Context) (*carddav.AddressBook, error) {
	p := ctx.Value(addressBookPathKey).(string)
	return &carddav.AddressBook{
		Path:                 p,
		Name:                 "My contacts",
		Description:          "Default address book",
		MaxResourceSize:      1024,
		SupportedAddressData: nil,
	}, nil
}

func (*Backend) GetAddressObject(ctx context.Context, path string, req *carddav.AddressDataRequest) (*carddav.AddressObject, error) {
	if path == alicePath {
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
}

func (b *Backend) ListAddressObjects(ctx context.Context, req *carddav.AddressDataRequest) ([]carddav.AddressObject, error) {
	alice, err := b.GetAddressObject(ctx, alicePath, req)
	if err != nil {
		return nil, err
	}

	return []carddav.AddressObject{*alice}, nil
}

func (*Backend) QueryAddressObjects(ctx context.Context, query *carddav.AddressBookQuery) ([]carddav.AddressObject, error) {
	panic("TODO: implement")
}

func (*Backend) PutAddressObject(ctx context.Context, path string, card vcard.Card, opts *carddav.PutAddressObjectOptions) (loc string, err error) {
	panic("TODO: implement")
}

func (*Backend) DeleteAddressObject(ctx context.Context, path string) error {
	panic("TODO: implement")
}
