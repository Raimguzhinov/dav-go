package carddav

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/emersion/go-vcard"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/carddav"
)

type Backend struct {
	repo Repository
}

func New(repository Repository) (*Backend, error) {
	return &Backend{
		repo: repository,
	}, nil
}

// type contextKey string
var (
	aliceData = `BEGIN:VCARD
VERSION:4.0
UID:urn:uuid:4fbe8971-0bc3-424c-9c26-36c3e1eff6b1
FN;PID=1.1:Alice Gopher
N:Gopher;Alice;;;
EMAIL;PID=1.1:alice@example.com
CLIENTPIDMAP:1;urn:uuid:53e374d9-337e-4727-8803-a1e9c14e0551
END:VCARD`
	//	alicePath = "urn:uuid:4fbe8971-0bc3-424c-9c26-36c3e1eff6b1.vcf"
	//
	//	currentUserPrincipalKey = contextKey("test:currentUserPrincipal")
	//	homeSetPathKey          = contextKey("test:homeSetPath")
	//	addressBookPathKey      = contextKey("test:addressBookPath")
)

func (*Backend) CurrentUserPrincipal(ctx context.Context) (string, error) {
	return "/admin/", nil
}

func (*Backend) AddressbookHomeSetPath(ctx context.Context) (string, error) {
	return "/admin/contacts/", nil
}

func (*Backend) AddressBook(ctx context.Context) (*carddav.AddressBook, error) {
	datatype := make([]carddav.AddressDataType, 0)
	datatype = append(datatype, carddav.AddressDataType{
		ContentType: "text/vcard",
		Version:     "3.0",
	})
	//p := ctx.Value(addressBookPathKey).(string)
	return &carddav.AddressBook{
		Path:                 path.Join("admin", "contacts", strconv.Itoa(1)),
		Name:                 "My contacts",
		Description:          "Default address book",
		MaxResourceSize:      1024,
		SupportedAddressData: datatype,
	}, nil
	//panic("TODO: implement")
}

func (*Backend) GetAddressObject(ctx context.Context, path string, req *carddav.AddressDataRequest) (*carddav.AddressObject, error) {
	if path == "/admin/contacts/alice/" {
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
	//panic("TODO: implement")
}

func (b *Backend) ListAddressObjects(ctx context.Context, req *carddav.AddressDataRequest) ([]carddav.AddressObject, error) {
	alice, err := b.GetAddressObject(ctx, "/admin/contacts/alice/", req)
	if err != nil {
		return nil, err
	}
	return []carddav.AddressObject{*alice}, nil
	//panic("TODO: implement")
}

func (b *Backend) QueryAddressObjects(ctx context.Context, query *carddav.AddressBookQuery) ([]carddav.AddressObject, error) {
	panic("TODO: implement")
}

func (b *Backend) PutAddressObject(ctx context.Context, path string, card vcard.Card, opts *carddav.PutAddressObjectOptions) (loc string, err error) {
	//b.PutAddressObject(ctx)
	panic("TODO: implement")
}

func (b *Backend) DeleteAddressObject(ctx context.Context, path string) error {
	panic("TODO: implement")
}
