package carddav_db

import (
	"context"
	"path"
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

func (r *repository) CreateFolder(ctx context.Context, homeSetPath string, addressbook *carddav.AddressBook) error {
	var f folder

	abUid, err := uuid.Parse(path.Clean(strings.TrimPrefix(addressbook.Path, homeSetPath)))
	if err != nil {
		r.logger.Error("postgres.CreateFolder", logger.Err(err))
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
		r.logger.Error("postgres.FindFolder", logger.Err(err))
		err = r.client.ToPgErr(err)
		return nil, err
	}

	var f folder
	addressbooks := make([]carddav.AddressBook, 0)

	for rows.Next() {
		var addressbook carddav.AddressBook

		err = rows.Scan(&f.Uid, &addressbook.Name, &addressbook.Description, &addressbook.MaxResourceSize, &f.Types)
		if err != nil {
			r.logger.Error("postgres.FindFolder", logger.Err(err))
			err = r.client.ToPgErr(err)
			return nil, err
		}

		addressbook.Path = path.Join(homeSetPath, f.Uid.String()) + "/"
		addressbook.SupportedAddressData, err = f.ParseTypes()
		if err != nil {
			r.logger.Error("postgres.FindFolder", logger.Err(err))
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

func (r *repository) PutAddressObject(ctx context.Context, homeSetPath string, object *carddav.AddressObject, opts *carddav.PutAddressObjectOptions) error {
	cf, err := fromAddressObject(object, homeSetPath)
	if err != nil {
		r.logger.Error("postgres.PutAddressObject", logger.Err(err))
		return err
	}

	_, err = r.client.Pool.Exec(ctx, `
		INSERT INTO carddav.card_file
		(
			 uid,
			 addressbook_folder_uid,
			 file_name,
			 etag,
			 created_at,
			 modified_at,
			 version,
			 formatted_name,
			 family_name,
			 given_name,
			 additional_names,
			 honorific_prefix,
			 honorific_suffix,
			 product,
			 kind,
			 nickname,
			 photo,
			 photo_media_type,
			 logo,
			 logo_media_type,
			 sound,
			 sound_media_type,
			 birthday,
			 anniversary,
			 gender,
			 revision_at,
			 language,
			 timezone,
			 geo,
			 title,
			 role,
			 organization_uid,
			 categories,
			 note
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34)
	`, &cf.UID, &cf.AddressbookFolderUID, &cf.FileName, &cf.Etag, &cf.CreatedAt, &cf.ModifiedAt, &cf.Version, &cf.FormattedName, &cf.FamilyName, &cf.GivenName, &cf.AdditionalNames, &cf.HonorificPrefix, &cf.HonorificSuffix, &cf.Product, &cf.Kind,
		&cf.Nickname, &cf.Photo, &cf.PhotoMediaType, &cf.Logo, &cf.LogoMediaType, &cf.Sound, &cf.SoundMediaType, &cf.Birthday, &cf.Anniversary, &cf.Gender,
		&cf.RevisionAt, &cf.Language, &cf.Timezone, &cf.Geo, &cf.Title, &cf.Role, &cf.OrganizationUID, &cf.Categories, &cf.Note)
	if err != nil {
		err = r.client.ToPgErr(err)
		r.logger.Error("postgres.PutAddressObject", logger.Err(err))
		return err
	}

	return nil
}

func (r *repository) FindAddressObjects(ctx context.Context, homeSetPath, abUIDstring string) ([]carddav.AddressObject, error) {
	abUID, err := uuid.Parse(abUIDstring)
	if err != nil {
		r.logger.Error("postgres.FindAddressObjects", logger.Err(err))
		return nil, err
	}

	rows, err := r.client.Pool.Query(ctx, `
		SELECT
			uid,
			addressbook_folder_uid,
			file_name,
			etag,
			created_at,
			modified_at,
			version,
			formatted_name,
			family_name,
			given_name,
			additional_names,
			honorific_prefix,
			honorific_suffix,
			product,
			kind,
			nickname,
			photo,
			photo_media_type,
			logo,
			logo_media_type,
			sound,
			sound_media_type,
			birthday,
			anniversary,
			gender,
			revision_at,
			language,
			timezone,
			geo,
			title,
			role,
			organization_uid,
			categories,
			note
		FROM
			carddav.card_file c
		WHERE
		    c.addressbook_folder_uid = $1
		GROUP BY
			c.uid, c.addressbook_folder_uid
		ORDER BY
			c.uid
		`, abUID)
	if err != nil {
		r.logger.Error("postgres.FindAddressObjects", logger.Err(err))
		err = r.client.ToPgErr(err)
		return nil, err
	}

	addressObjects := make([]carddav.AddressObject, 0)

	for rows.Next() {
		var ao carddav.AddressObject
		var cf cardFile
		err = rows.Scan(&cf.UID, &cf.AddressbookFolderUID, &cf.FileName, &cf.Etag, &cf.CreatedAt, &cf.ModifiedAt, &cf.Version, &cf.FormattedName, &cf.FamilyName, &cf.GivenName, &cf.AdditionalNames, &cf.HonorificPrefix, &cf.HonorificSuffix, &cf.Product, &cf.Kind,
			&cf.Nickname, &cf.Photo, &cf.PhotoMediaType, &cf.Logo, &cf.LogoMediaType, &cf.Sound, &cf.SoundMediaType, &cf.Birthday, &cf.Anniversary, &cf.Gender,
			&cf.RevisionAt, &cf.Language, &cf.Timezone, &cf.Geo, &cf.Title, &cf.Role, &cf.OrganizationUID, &cf.Categories, &cf.Note)
		if err != nil {
			r.logger.Error("postgres.FindAddressObjects", logger.Err(err))
			err = r.client.ToPgErr(err)
			return nil, err
		}

		err = cf.toAddressObject(&ao)
		ao.Path = path.Join(homeSetPath, ao.Path)
		if err != nil {
			r.logger.Error("postgres.FindAddressObjects", logger.Err(err))
			return nil, err
		}
		addressObjects = append(addressObjects, ao)
	}

	return addressObjects, nil
}

func (r *repository) DeleteAddressObject(ctx context.Context, path string) error {
	//panic("TODO")
	return nil
}

func (r *repository) GetFolderAccess(ctx context.Context, addressbook *carddav.AddressBook) ([]string, error) {

	return nil, nil
}
