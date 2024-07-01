package db

import (
	"encoding/base64"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/ceres919/go-webdav/carddav"
	"github.com/emersion/go-vcard"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

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

type cardEmail struct {
	ID              pgtype.Uint32 `json:"id"`
	CardFileUID     pgtype.UUID   `json:"card_file_uid"`
	Type            pgtype.Text   `json:"type"`
	Email           pgtype.Text   `json:"email"`
	PreferenceLevel pgtype.Uint32 `json:"preference_level"`
	SortIndex       pgtype.Int8   `json:"sort_index"`
}

type cardTelephone struct {
	ID              pgtype.Uint32 `json:"id"`
	CardFileUID     pgtype.UUID   `json:"card_file_uid"`
	Type            pgtype.Text   `json:"type"`
	Telephone       pgtype.Text   `json:"telephone"`
	PreferenceLevel pgtype.Uint32 `json:"preference_level"`
	SortIndex       pgtype.Int8   `json:"sort_index"`
}

type cardURL struct {
	ID              pgtype.Uint32 `json:"id"`
	CardFileUID     pgtype.UUID   `json:"card_file_uid"`
	Type            pgtype.Text   `json:"type"`
	URL             pgtype.Text   `json:"url"`
	PreferenceLevel pgtype.Uint32 `json:"preference_level"`
	SortIndex       pgtype.Int8   `json:"sort_index"`
}

type cardInstantMessenger struct {
	ID               pgtype.Uint32 `json:"id"`
	CardFileUID      pgtype.UUID   `json:"card_file_uid"`
	Type             pgtype.Text   `json:"type"`
	InstantMessenger pgtype.Text   `json:"instant_messenger"`
	PreferenceLevel  pgtype.Uint32 `json:"preference_level"`
	SortIndex        pgtype.Int8   `json:"sort_index"`
}

type customeProperty struct {
}

type cardFile struct {
	UID                  pgtype.UUID          `json:"uid"`
	AddressbookFolderUID pgtype.UUID          `json:"addressbook_folder_uid"`
	FileName             pgtype.Text          `json:"file_name"`
	Etag                 pgtype.Text          `json:"etag"`
	CreatedAt            pgtype.Timestamp     `json:"created_at"`
	ModifiedAt           pgtype.Timestamp     `json:"modified_at"`
	Version              pgtype.Text          `json:"version"`
	FormattedName        pgtype.Text          `json:"formatted_name"`
	FamilyName           pgtype.Text          `json:"family_name"`
	GivenName            pgtype.Text          `json:"given_name"`
	AdditionalNames      pgtype.Text          `json:"additional_names"`
	HonorificPrefix      pgtype.Text          `json:"honorific_prefix"`
	HonorificSuffix      pgtype.Text          `json:"honorific_suffix"`
	Product              pgtype.Text          `json:"product,omitempty"`
	Kind                 pgtype.Text          `json:"kind,omitempty"`
	Nickname             pgtype.Text          `json:"nickname,omitempty"`
	Photo                pgtype.PreallocBytes `json:"photo,omitempty"`
	PhotoMediaType       pgtype.Text          `json:"photo_media_type,omitempty"`
	Logo                 pgtype.PreallocBytes `json:"logo,omitempty"`
	LogoMediaType        pgtype.Text          `json:"logo_media_type,omitempty"`
	Sound                pgtype.PreallocBytes `json:"sound,omitempty"`
	SoundMediaType       pgtype.Text          `json:"sound_media_type,omitempty"`
	Birthday             pgtype.Date          `json:"birthday,omitempty"`
	Anniversary          pgtype.Date          `json:"anniversary,omitempty"`
	Gender               pgtype.Text          `json:"gender,omitempty"`
	RevisionAt           pgtype.Timestamp     `json:"revision_at,omitempty"`
	Language             pgtype.Text          `json:"language,omitempty"`
	Timezone             pgtype.Text          `json:"timezone,omitempty"`
	Geo                  pgtype.Point         `json:"geo,omitempty"`
	Title                pgtype.Text          `json:"title,omitempty"`
	Role                 pgtype.Text          `json:"role,omitempty"`
	OrganizationUID      pgtype.UUID          `json:"organization_uid,omitempty"`
	Categories           pgtype.Text          `json:"categories,omitempty"`
	Note                 pgtype.Text          `json:"note,omitempty"`
	//Email
	//Telephone
	//URL
	//InstantMessenger
}

func (c *cardFile) toAddressObject(obj *carddav.AddressObject) error {

	// TODO email, tel, url, impp

	abUID, err := uuid.FromBytes(c.AddressbookFolderUID.Bytes[:])
	if err != nil {
		return err
	}
	obj.Path = path.Join(abUID.String(), c.FileName.String)
	obj.ETag = c.Etag.String
	obj.ModTime = c.ModifiedAt.Time
	obj.Card = vcard.Card{}

	uid, err := uuid.FromBytes(c.UID.Bytes[:])
	if err != nil {
		return err
	}
	setVcardValue(&obj.Card, vcard.FieldUID, uid.String())
	setVcardValue(&obj.Card, vcard.FieldVersion, c.Version.String)
	setVcardValue(&obj.Card, vcard.FieldFormattedName, c.FormattedName.String)
	if c.FamilyName.Valid || c.GivenName.Valid || c.AdditionalNames.Valid || c.HonorificPrefix.Valid || c.HonorificSuffix.Valid {
		name := strings.Join([]string{c.FamilyName.String, c.GivenName.String, c.AdditionalNames.String, c.HonorificPrefix.String, c.HonorificSuffix.String}, ";")
		setVcardValue(&obj.Card, vcard.FieldName, name)
	}
	setVcardValue(&obj.Card, vcard.FieldProductID, c.Product.String)
	setVcardValue(&obj.Card, vcard.FieldKind, c.Kind.String)
	setVcardValue(&obj.Card, vcard.FieldNickname, c.Nickname.String)
	if c.Photo != nil {
		setVcardValue(&obj.Card, vcard.FieldPhoto, strings.Join([]string{c.PhotoMediaType.String, base64.StdEncoding.EncodeToString(c.Photo[:])}, ","))
	}
	if c.Logo != nil {
		setVcardValue(&obj.Card, vcard.FieldLogo, strings.Join([]string{c.LogoMediaType.String, base64.StdEncoding.EncodeToString(c.Logo[:])}, ","))
	}
	if c.Sound != nil {
		setVcardValue(&obj.Card, vcard.FieldSound, strings.Join([]string{c.SoundMediaType.String, base64.StdEncoding.EncodeToString(c.Sound[:])}, ","))
	}
	setVcardValue(&obj.Card, vcard.FieldBirthday, c.Birthday.Time.Format("20060102"))
	setVcardValue(&obj.Card, vcard.FieldAnniversary, c.Anniversary.Time.Format("20060102"))
	setVcardValue(&obj.Card, vcard.FieldGender, c.Gender.String)
	//setVcardValue(&obj.Card, vcard.FieldRevision, c.RevisionAt.Time.String())
	setVcardValue(&obj.Card, vcard.FieldLanguage, c.Language.String)
	setVcardValue(&obj.Card, vcard.FieldTimezone, c.Timezone.String)
	if c.Geo.Valid {
		setVcardValue(&obj.Card, vcard.FieldGeolocation, strings.Join([]string{fmt.Sprintf("%f", c.Geo.P.X), fmt.Sprintf("%f", c.Geo.P.Y)}, ","))
	}
	setVcardValue(&obj.Card, vcard.FieldTitle, c.Title.String)
	setVcardValue(&obj.Card, vcard.FieldRole, c.Role.String)
	setVcardValue(&obj.Card, vcard.FieldCategories, c.Categories.String)
	setVcardValue(&obj.Card, vcard.FieldNote, c.Note.String)
	return nil
}

func setVcardValue(c *vcard.Card, propName, val string) {
	if val == "" {
		return
	}
	c.AddValue(propName, val)
	return
}

func fromAddressObject(obj *carddav.AddressObject, homeSetPath string) (*cardFile, error) {
	splitPath := strings.Split(strings.TrimPrefix(obj.Path, homeSetPath), "/")
	abUID, err := uuid.Parse(splitPath[0])
	if err != nil {
		return nil, err
	}

	var names [5]pgtype.Text
	f := obj.Card.Get(vcard.FieldName)
	if f == nil {
		for i := range names {
			names[i] = pgtype.Text{Valid: false}
		}
	}
	splitNames := strings.Split(f.Value, ";")
	for i := range names {
		names[i] = pgtype.Text{
			String: splitNames[i],
			Valid:  true,
		}
	}

	photo, photoType := getMediaValue(&obj.Card, vcard.FieldPhoto)
	logo, logoType := getMediaValue(&obj.Card, vcard.FieldLogo)
	sound, soundType := getMediaValue(&obj.Card, vcard.FieldSound)

	// TODO email, tel, url, impp, organization, CreatedAt

	return &cardFile{
		UID: getUIDValue(&obj.Card, vcard.FieldUID),
		AddressbookFolderUID: pgtype.UUID{
			Bytes: abUID,
			Valid: true,
		},
		FileName: pgtype.Text{
			String: splitPath[1],
			Valid:  true,
		},
		Etag: pgtype.Text{
			String: obj.ETag,
			Valid:  true,
		},
		CreatedAt: pgtype.Timestamp{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		ModifiedAt: pgtype.Timestamp{
			Time:  time.Now().UTC(),
			Valid: true,
		},
		Version:         getTextValue(&obj.Card, vcard.FieldVersion),
		FormattedName:   getTextValue(&obj.Card, vcard.FieldFormattedName),
		FamilyName:      names[0],
		GivenName:       names[1],
		AdditionalNames: names[2],
		HonorificPrefix: names[3],
		HonorificSuffix: names[4],
		Product:         getTextValue(&obj.Card, vcard.FieldProductID),
		Kind:            getTextValue(&obj.Card, vcard.FieldKind),
		Nickname:        getTextValue(&obj.Card, vcard.FieldNickname),
		Photo:           photo,
		PhotoMediaType:  photoType,
		Logo:            logo,
		LogoMediaType:   logoType,
		Sound:           sound,
		SoundMediaType:  soundType,
		Birthday:        getDateValue(&obj.Card, vcard.FieldBirthday),
		Anniversary:     getDateValue(&obj.Card, vcard.FieldAnniversary),
		Gender:          getTextValue(&obj.Card, vcard.FieldGender),
		RevisionAt:      getTimestampValue(&obj.Card, vcard.FieldRevision),
		Language:        getTextValue(&obj.Card, vcard.FieldLanguage),
		Timezone:        getTextValue(&obj.Card, vcard.FieldTimezone),
		Geo:             getPointValue(&obj.Card, vcard.FieldGeolocation),
		Title:           getTextValue(&obj.Card, vcard.FieldTitle),
		Role:            getTextValue(&obj.Card, vcard.FieldRole),
		OrganizationUID: pgtype.UUID{Valid: false},
		Categories:      getTextValue(&obj.Card, vcard.FieldCategories),
		Note:            getTextValue(&obj.Card, vcard.FieldNote),
	}, nil
}

func getUIDValue(c *vcard.Card, prop string) pgtype.UUID {
	f := c.Get(prop)
	if f == nil {
		return pgtype.UUID{Valid: false}
	}

	uid, err := uuid.Parse(f.Value)
	if err != nil {
		return pgtype.UUID{Valid: false}
	}

	return pgtype.UUID{
		Bytes: uid,
		Valid: true,
	}
}

func getTextValue(c *vcard.Card, prop string) pgtype.Text {
	f := c.Get(prop)
	if f == nil {
		return pgtype.Text{Valid: false}
	}

	return pgtype.Text{
		String: f.Value,
		Valid:  true,
	}
}

func getTimestampValue(c *vcard.Card, prop string) pgtype.Timestamp {
	f := c.Get(prop)
	if f == nil {
		return pgtype.Timestamp{Valid: false}
	}

	var t pgtype.Timestamp
	err := t.Scan(f.Value)
	if err != nil {
		return pgtype.Timestamp{Valid: false}
	}

	return t
}

func getDateValue(c *vcard.Card, prop string) pgtype.Date {
	f := c.Get(prop)
	if f == nil {
		return pgtype.Date{Valid: false}
	}

	t, err := time.Parse("20060102", f.Value)
	if err != nil {
		return pgtype.Date{Valid: false}
	}

	return pgtype.Date{
		Time:  t,
		Valid: true,
	}
}

func getPointValue(c *vcard.Card, prop string) pgtype.Point {
	f := c.Get(prop)
	if f == nil {
		return pgtype.Point{Valid: false}
	}

	var p pgtype.Point
	err := p.Scan(f.Value)
	if err != nil {
		return pgtype.Point{Valid: false}
	}

	return p
}

func getMediaValue(c *vcard.Card, prop string) (pgtype.PreallocBytes, pgtype.Text) {
	f := c.Get(prop)
	if f == nil {
		return nil, pgtype.Text{Valid: false}
	}

	split := strings.Split(f.Value, ",")
	b, err := base64.StdEncoding.DecodeString(split[1])
	if err != nil {
		return nil, pgtype.Text{Valid: false}
	}
	return b, pgtype.Text{
		String: split[0],
		Valid:  true,
	}
}
