package carddav_db

import (
	"context"
	"path"
	"regexp"
	"strconv"

	backend "github.com/Raimguhinov/dav-go/internal/carddav"
	"github.com/Raimguhinov/dav-go/pkg/logger"
	"github.com/Raimguhinov/dav-go/pkg/postgres"
	"github.com/ceres919/go-webdav/carddav"
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
	ID    int      `json:"id"`
	Types []string `json:"types"`
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

	err := r.client.Pool.QueryRow(ctx, `
		INSERT INTO carddav.addressbook_folder
			(name, description, max_resource_size, supported_address_data)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, addressbook.Name, addressbook.Description, addressbook.MaxResourceSize, addressbook.SupportedAddressData).Scan(&f.ID)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.CreateFolder", logger.Err(err))
		return err
	}
	addressbook.Path = path.Join(homeSetPath, strconv.Itoa(f.ID), "/")
	return nil
}

func (r *repository) FindFolders(ctx context.Context) ([]carddav.AddressBook, error) {
	rows, err := r.client.Pool.Query(ctx, `
		SELECT
			f.id,
			f.name,
    		COALESCE(f.description, '') as description,
			f.max_resource_size AS size,
			array_agg(f.supported_address_data) AS types
		FROM
			carddav.addressbook_folder f
		GROUP BY
			f.id, f.name, f.description
		ORDER BY
			f.id; 
		`)
	if err != nil {
		err = r.client.ToPgErr(err)
		return nil, err
	}

	var f folder
	addressbooks := make([]carddav.AddressBook, 0)

	for rows.Next() {
		var addressbook carddav.AddressBook

		err = rows.Scan(&f.ID, &addressbook.Name, &addressbook.Description, &addressbook.MaxResourceSize, &f.Types)
		if err != nil {
			err = r.client.ToPgErr(err)
			return nil, err
		}

		addressbook.Path = strconv.Itoa(f.ID)
		addressbook.SupportedAddressData, err = f.ParseTypes()
		if err != nil {
			return nil, err
		}
		addressbooks = append(addressbooks, addressbook)
	}

	return addressbooks, nil
}

func (r *repository) DeleteFolder(ctx context.Context, addressbook *carddav.AddressBook) error {
	panic("TODO")
}

func (r *repository) PutObject(ctx context.Context, uid string, object *carddav.AddressObject, opts *carddav.PutAddressObjectOptions) (string, error) {
	panic("TODO")
}

func (r *repository) CreateContact(ctx context.Context, addressbook *carddav.AddressBook) error {
	panic("TODO")
}

func (r *repository) FindContacts(ctx context.Context) ([]carddav.AddressObject, error) {
	panic("TODO")
}
