package carddav_db

import (
	"context"
	"regexp"
	"strings"

	backend "github.com/Raimguhinov/dav-go/internal/carddav"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/ceres919/go-webdav/carddav"
	"github.com/google/uuid"
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

type folder struct {
	Uid   uuid.UUID `json:"uid"`
	Types []string  `json:"types"`
}

func (f *folder) ParseTypes() ([]carddav.AddressDataType, error) {
	supTypes := make([]carddav.AddressDataType, 0)
	re := regexp.MustCompile(`\((.*),(.*)\)`)

	for _, t := range f.Types {
		var supType carddav.AddressDataType
		result := re.FindStringSubmatch(t)
		supType.ContentType = result[0]
		supType.Version = result[1]
	}
	return supTypes, nil
}

func (r *repository) CreateFolder(ctx context.Context, homeSetPath string, addressbook *carddav.AddressBook) error {
	var f folder

	abUid, err := uuid.Parse(strings.TrimSuffix(strings.TrimPrefix(addressbook.Path, homeSetPath), "/"))
	if err != nil {
		return err
	}

	err = r.client.Pool.QueryRow(ctx, `
		INSERT INTO carddav.addressbook_folder
			(uid, name, description)
		VALUES ($1, $2, $3)
		RETURNING uid
	`, abUid, addressbook.Name, addressbook.Description).Scan(&f.Uid)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.CreateFolder", logger.Err(err))
		return err
	}
	return nil
}

func (r *repository) FindFolders(ctx context.Context, homeSetPath string) ([]carddav.AddressBook, error) {
	rows, err := r.client.Pool.Query(ctx, `
		SELECT
			f.uid,
			f.name,
			COALESCE(f.description, '') as description,
			f.max_resource_size AS size,
			array_agg(f.supported_address_data) AS types
		FROM
			carddav.addressbook_folder f
		GROUP BY
			f.uid, f.name
		ORDER BY
			f.uid
		`)
	if err != nil {
		err = r.client.ToPgErr(err)
		return nil, err
	}

	var f folder
	addressbooks := make([]carddav.AddressBook, 0)

	for rows.Next() {
		var addressbook carddav.AddressBook

		err = rows.Scan(&f.Uid, &addressbook.Name, &addressbook.Description, &addressbook.MaxResourceSize, &f.Types)
		if err != nil {
			err = r.client.ToPgErr(err)
			return nil, err
		}

		addressbook.Path = f.Uid.String()
		addressbook.SupportedAddressData, err = f.ParseTypes()
		if err != nil {
			return nil, err
		}
		addressbooks = append(addressbooks, addressbook)
	}

	return addressbooks, nil
}

func (r *repository) DeleteFolder(ctx context.Context, addressbook *carddav.AddressBook) error {
	//panic("TODO")
	return nil
}

func (r *repository) PutAddressObject(ctx context.Context, object *carddav.AddressObject, opts *carddav.PutAddressObjectOptions) (string, error) {
	//panic("TODO")
	return "", nil
}

func (r *repository) FindAddressObjects(ctx context.Context) ([]carddav.AddressObject, error) {
	//panic("TODO")
	return nil, nil
}

func (r *repository) DeleteAddressObject(ctx context.Context, path string) error {
	//panic("TODO")
	return nil
}

func (r *repository) GetFolderAccess(ctx context.Context, addressbook *carddav.AddressBook) ([]string, error) {
	//panic("TODO")
	return nil, nil
}
