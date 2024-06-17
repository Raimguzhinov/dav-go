BEGIN;

CREATE SCHEMA IF NOT EXISTS carddav;

CREATE TYPE carddav.address_data_type AS
(
    content_type TEXT,
    type_version TEXT
);

CREATE TABLE IF NOT EXISTS carddav.addressbook_folder
(
    uid                    UUID PRIMARY KEY,
    name                   VARCHAR(50) NOT NULL,
    description            TEXT                        DEFAULT NULL,
    max_resource_size      INT                         DEFAULT 4096,
    supported_address_data carddav.address_data_type[] DEFAULT ARRAY [ROW ('text/vcard', '3.0'), ROW ('text/vcard', '4.0')]::carddav.address_data_type[]
);

CREATE TABLE IF NOT EXISTS carddav.access
(
    id                     BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    addressbook_folder_uid UUID        NOT NULL,
    user_id                VARCHAR(50) NOT NULL,
    owner                  BIT         NOT NULL,
    read                   BIT         NOT NULL,
    write                  BIT         NOT NULL,
    CONSTRAINT fk_addressbook_folder FOREIGN KEY (addressbook_folder_uid) REFERENCES carddav.addressbook_folder (uid)
);

CREATE TABLE IF NOT EXISTS carddav.card_file
(
    uid                    UUID PRIMARY KEY,
    addressbook_folder_uid UUID         NOT NULL,
    file_name              VARCHAR(255) NOT NULL,
    etag                   TIMESTAMP    NOT NULL,
    created_at             TIMESTAMP    NOT NULL,
    modified_at            TIMESTAMP    NOT NULL,
    version                VARCHAR(5)   NOT NULL,
    formatted_name         VARCHAR(100) NOT NULL,
    family_name            VARCHAR(100) NOT NULL,
    given_name             VARCHAR(100) NOT NULL,
    additional_names       VARCHAR(100) NOT NULL,
    honorific_prefix       VARCHAR(10)  NOT NULL,
    honorific_suffix       VARCHAR(10)  NOT NULL,
    product                VARCHAR(100),
    kind                   VARCHAR(10),
    nickname               VARCHAR(50),
    photo                  BYTEA,
    photo_media_type       VARCHAR(10),
    logo                   BYTEA,
    logo_media_type        VARCHAR(10),
    sound                  BYTEA,
    sound_media_type       VARCHAR(10),
    birthday               DATE,
    anniversary            DATE,
    gender                 VARCHAR(1),
    revision_at            TIMESTAMP,
    sort_string            VARCHAR(100),
    language               VARCHAR(50),
    timezone               VARCHAR(50),
    geo                    POINT,
    title                  TEXT,
    role                   VARCHAR(50),
    org_name               VARCHAR(100),
    org_unit               VARCHAR(50),
    categories             VARCHAR(50),
    note                   TEXT,
    classification         VARCHAR(50),
    CONSTRAINT fk_addressbook_folder FOREIGN KEY (addressbook_folder_uid) REFERENCES carddav.addressbook_folder (uid)
);

CREATE INDEX ON carddav.card_file USING GIST (geo);

CREATE TABLE IF NOT EXISTS carddav.custom_property
(
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id       BIGINT       NOT NULL,
    card_file_uid   UUID         NOT NULL,
    client_app_name VARCHAR(50),
    prop_name       VARCHAR(50)  NOT NULL,
    parameter_name  VARCHAR(50),
    value           VARCHAR(512) NOT NULL,
    sort_index      INT,
    CONSTRAINT fk_card_file_uid FOREIGN KEY (card_file_uid) REFERENCES carddav.card_file (uid)
);

CREATE TABLE IF NOT EXISTS carddav.email
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    card_file_uid    UUID         NOT NULL,
    type             VARCHAR(20),
    email            VARCHAR(100) NOT NULL,
    preference_level SMALLINT,
    sort_index       INT,
    CONSTRAINT fk_card_file_uid FOREIGN KEY (card_file_uid) REFERENCES carddav.card_file (uid)
);

CREATE TABLE IF NOT EXISTS carddav.telephone
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    card_file_uid    UUID        NOT NULL,
    type             VARCHAR(20),
    telephone        VARCHAR(20) NOT NULL,
    preference_level SMALLINT,
    sort_index       INT,
    CONSTRAINT fk_card_file_uid FOREIGN KEY (card_file_uid) REFERENCES carddav.card_file (uid)
);

CREATE TABLE IF NOT EXISTS carddav.url
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    card_file_uid    UUID NOT NULL,
    type             VARCHAR(20),
    url              TEXT NOT NULL,
    preference_level SMALLINT,
    sort_index       INT,
    CONSTRAINT fk_card_file_uid FOREIGN KEY (card_file_uid) REFERENCES carddav.card_file (uid)
);

CREATE TABLE IF NOT EXISTS carddav.instant_messenger
(
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    card_file_uid     UUID NOT NULL,
    type              VARCHAR(20),
    instant_messenger TEXT NOT NULL,
    preference_level  SMALLINT,
    sort_index        INT,
    CONSTRAINT fk_card_file_uid FOREIGN KEY (card_file_uid) REFERENCES carddav.card_file (uid)
);

CREATE TABLE IF NOT EXISTS carddav.address
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    card_file_uid    UUID NOT NULL,
    type             VARCHAR(20),
    po_box           VARCHAR(10),
    apartment_number VARCHAR(10),
    street           VARCHAR(50),
    locality         VARCHAR(50),
    region           VARCHAR(50),
    postal_code      VARCHAR(10),
    country          VARCHAR(20),
    preference_level SMALLINT,
    label            VARCHAR(20),
    geo              POINT,
    timezone         VARCHAR(50),
    sort_index       INT,
    CONSTRAINT fk_card_file_uid FOREIGN KEY (card_file_uid) REFERENCES carddav.card_file (uid)
);

CREATE INDEX ON carddav.address USING GIST (geo);

COMMIT;
