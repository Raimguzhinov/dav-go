package carddav

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/Raimguhinov/dav-go/internal/usecase/etag"
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
		Path:        path.Join(homeSetPath, uid.String()) + "/",
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

	return addressbooks, nil
}

func (b *backend) GetAddressBook(ctx context.Context, urlPath string) (*carddav.AddressBook, error) {
	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return nil, err
	}

	addressbooks, err := b.repo.FindFolders(ctx, homeSetPath)
	if err != nil {
		return nil, err
	}

	for _, addressbook := range addressbooks {
		if addressbook.Path == urlPath {
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
	//TODO
	//addressbook, err := b.GetAddressBook(ctx, path)
	//if err != nil {
	//	return err
	//}
	//err = b.repo.DeleteFolder(ctx, addressbook)
	//if err != nil {
	//	return err
	//}
	return nil
}

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

func (b *backend) QueryAddressObjects(ctx context.Context, urlPath string, query *carddav.AddressBookQuery) ([]carddav.AddressObject, error) {
	//TODO
	return nil, nil
}

func (b *backend) PutAddressObject(ctx context.Context, urlPath string, card vcard.Card, opts *carddav.PutAddressObjectOptions) (*carddav.AddressObject, error) {
	homeSetPath, err := b.AddressBookHomeSetPath(ctx)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = vcard.NewEncoder(bufio.NewWriter(&buf)).Encode(card)
	eTag, err := etag.FromData(buf.Bytes())

	ao := carddav.AddressObject{
		Path:    urlPath,
		ModTime: time.Now().UTC(),
		ETag:    eTag,
		Card:    card,
	}

	err = b.repo.PutAddressObject(ctx, homeSetPath, &ao, opts)
	if err != nil {
		return nil, err
	}
	return &ao, nil
}

func (b *backend) DeleteAddressObject(ctx context.Context, urlPath string) error {
	//TODO
	return nil
}

func (b *backend) GetCurrentUserAddressBookPrivilege(ctx context.Context, ab *carddav.AddressBook) ([]string, error) {
	return []string{"all", "read", "write", "write-properties", "write-content", "unlock", "bind", "unbind", "write-acl", "read-acl", "read-current-user-privilege-set"}, nil
}
