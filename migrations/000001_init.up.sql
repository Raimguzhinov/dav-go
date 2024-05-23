BEGIN;

CREATE SCHEMA IF NOT EXISTS carddav;
CREATE SCHEMA IF NOT EXISTS caldav;

CREATE TABLE IF NOT EXISTS carddav.addressbook_folder
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(50) NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS carddav.addressbook_folder_property
(
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    addressbook_folder_id BIGINT       NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    namespace             VARCHAR(100) NOT NULL,
    prop_value            TEXT         NOT NULL,
    CONSTRAINT fk_addressbook_folder FOREIGN KEY (addressbook_folder_id) REFERENCES carddav.addressbook_folder (id)
);

CREATE TABLE IF NOT EXISTS carddav.access
(
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    addressbook_folder_id BIGINT      NOT NULL,
    user_id               VARCHAR(50) NOT NULL,
    owner                 BIT         NOT NULL,
    read                  BIT         NOT NULL,
    write                 BIT         NOT NULL,
    CONSTRAINT fk_addressbook_folder FOREIGN KEY (addressbook_folder_id) REFERENCES carddav.addressbook_folder (id)
);

CREATE TABLE IF NOT EXISTS carddav.card_file
(
    uid                   UUID PRIMARY KEY,
    addressbook_folder_id BIGINT       NOT NULL,
    file_name             VARCHAR(255) NOT NULL,
    etag                  TIMESTAMP    NOT NULL,
    created_at            TIMESTAMP    NOT NULL,
    modified_at           TIMESTAMP    NOT NULL,
    version               VARCHAR(5)   NOT NULL,
    formatted_name        VARCHAR(100) NOT NULL,
    family_name           VARCHAR(100) NOT NULL,
    given_name            VARCHAR(100) NOT NULL,
    additional_names      VARCHAR(100) NOT NULL,
    honorific_prefix      VARCHAR(10)  NOT NULL,
    honorific_suffix      VARCHAR(10)  NOT NULL,
    product               VARCHAR(100),
    kind                  VARCHAR(10),
    nickname              VARCHAR(50),
    photo                 BYTEA,
    photo_media_type      VARCHAR(10),
    logo                  BYTEA,
    logo_media_type       VARCHAR(10),
    sound                 BYTEA,
    sound_media_type      VARCHAR(10),
    birthday              DATE,
    anniversary           DATE,
    gender                VARCHAR(1),
    revision_at           TIMESTAMP,
    sort_string           VARCHAR(100),
    language              VARCHAR(50),
    timezone              VARCHAR(50),
    geo                   POINT,
    title                 TEXT,
    role                  VARCHAR(50),
    org_name              VARCHAR(100),
    org_unit              VARCHAR(50),
    categories            VARCHAR(50),
    note                  TEXT,
    classification        VARCHAR(50),
    CONSTRAINT fk_addressbook_folder FOREIGN KEY (addressbook_folder_id) REFERENCES carddav.addressbook_folder (id)
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
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    card_file_uid     UUID NOT NULL,
    type              VARCHAR(20),
    po_box            VARCHAR(10),
    appartment_number VARCHAR(10),
    street            VARCHAR(50),
    locality          VARCHAR(50),
    region            VARCHAR(50),
    postal_code       VARCHAR(10),
    country           VARCHAR(20),
    preference_level  SMALLINT,
    label             VARCHAR(20),
    geo               POINT,
    timezone          VARCHAR(50),
    sort_index        INT,
    CONSTRAINT fk_card_file_uid FOREIGN KEY (card_file_uid) REFERENCES carddav.card_file (uid)
);

CREATE INDEX ON carddav.address USING GIST (geo);

CREATE TYPE caldav.calendar_type AS ENUM ('VEVENT', 'VTODO', 'VJOURNAL');

CREATE TABLE IF NOT EXISTS caldav.calendar_folder
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(50)          NOT NULL,
    type        caldav.calendar_type NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS caldav.calendar_folder_property
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_folder_id BIGINT       NOT NULL,
    name               VARCHAR(255) NOT NULL,
    namespace          VARCHAR(100) NOT NULL,
    prop_value         TEXT         NOT NULL,
    CONSTRAINT fk_calendar_folder FOREIGN KEY (calendar_folder_id) REFERENCES caldav.calendar_folder (id)
);

INSERT INTO caldav.calendar_folder (name, type)
VALUES ('calendars', 'VEVENT');
INSERT INTO caldav.calendar_folder (name, type)
VALUES ('todos', 'VTODO');
INSERT INTO caldav.calendar_folder (name, type)
VALUES ('journals', 'VJOURNAL');

CREATE TABLE IF NOT EXISTS caldav.access
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_folder_id BIGINT      NOT NULL,
    user_id            VARCHAR(50) NOT NULL,
    owner              BIT         NOT NULL,
    read               BIT         NOT NULL,
    write              BIT         NOT NULL,
    CONSTRAINT fk_calendar_folder FOREIGN KEY (calendar_folder_id) REFERENCES caldav.calendar_folder (id)
);

CREATE TABLE IF NOT EXISTS caldav.calendar_file
(
    uid                UUID PRIMARY KEY,
    calendar_folder_id BIGINT    NOT NULL,
    etag               TIMESTAMP NOT NULL,
    created_at         TIMESTAMP NOT NULL,
    modified_at        TIMESTAMP NOT NULL,
    CONSTRAINT fk_calendar_folder FOREIGN KEY (calendar_folder_id) REFERENCES caldav.calendar_folder (id)
);

CREATE TABLE IF NOT EXISTS caldav.custom_property
(
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id         BIGINT       NOT NULL,
    calendar_file_uid UUID         NOT NULL,
    prop_name         VARCHAR(50)  NOT NULL,
    parameter_name    VARCHAR(50),
    value             VARCHAR(512) NOT NULL,
    CONSTRAINT fk_calendar_file FOREIGN KEY (calendar_file_uid) REFERENCES caldav.calendar_file (uid)
);

CREATE TABLE IF NOT EXISTS caldav.event_component
(
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_file_uid     UUID NOT NULL,
    component_type        BIT  NOT NULL,
    date_timestamp        DATE NOT NULL,
    created_at            DATE,
    last_modified_at      DATE,
    summary               VARCHAR(512),
    description           TEXT,
    organizer_email       VARCHAR(255),
    organizer_common_name VARCHAR(50),
    start_date            DATE,
    start_timezone_id     VARCHAR(255),
    end_date              DATE,
    end_timezone_id       VARCHAR(255),
    duration              BIGINT,
    all_day               BIT,
    class                 VARCHAR(50),
    location              VARCHAR(255),
    priority              SMALLINT,
    sequence              INT,
    status                VARCHAR(50),
    categories            VARCHAR(255),
    event_transparency    BIT,
    todo_completed        DATE,
    todo_percent_complete SMALLINT,
    CONSTRAINT fk_calendar_file FOREIGN KEY (calendar_file_uid) REFERENCES caldav.calendar_file (uid)
);

CREATE TABLE IF NOT EXISTS caldav.attachment
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id BIGINT NOT NULL,
    uid                UUID   NOT NULL,
    media_type         VARCHAR(255),
    external_url       TEXT,
    content            BYTEA,
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES caldav.event_component (id)
);

CREATE TABLE IF NOT EXISTS caldav.attendee
(
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id   BIGINT NOT NULL,
    uid                  UUID   NOT NULL,
    email                VARCHAR(255),
    common_name          VARCHAR(50),
    directory_entry_ref  TEXT,
    language             VARCHAR(50),
    user_type            VARCHAR(15),
    sent_by              VARCHAR(50),
    delegated_from       VARCHAR(50),
    delegated_to         VARCHAR(50),
    rsvp                 BIT,
    participation_role   VARCHAR(15),
    participation_status VARCHAR(15),
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES caldav.event_component (id)
);

CREATE TABLE IF NOT EXISTS caldav.alarm
(
    id                        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id        BIGINT      NOT NULL,
    uid                       UUID        NOT NULL,
    action                    VARCHAR(15) NOT NULL,
    trigger_absolute_datetime DATE,
    trigger_relative_offset   BIGINT,
    trigger_related_start     BIT,
    summary                   VARCHAR(255),
    description               TEXT,
    duration                  BIGINT,
    repeat                    INT,
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES caldav.event_component (id)
);

CREATE TABLE IF NOT EXISTS caldav.recurrence
(
    id                         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id         BIGINT NOT NULL,
    recur_interval             INT,
    recur_until                DATE,
    recur_count                INT,
    recur_week_start           VARCHAR(50),
    recur_by_day_mask          BIT(7),
    recur_by_month_day_mask    VARCHAR(50),
    recur_by_set_pos           VARCHAR(50),
    recurrence_id_date         DATE,
    recurrence_id_timezone_id  VARCHAR(255),
    recurrence_this_and_future BIT,
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES caldav.event_component (id)
);

CREATE TABLE IF NOT EXISTS caldav.recurrence_exception
(
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    recurrence_id  BIGINT NOT NULL,
    uid            UUID   NOT NULL,
    exception_date DATE   NOT NULL,
    timezone_id    VARCHAR(255),
    all_day        BIT    NOT NULL,
    CONSTRAINT fk_recurrence FOREIGN KEY (recurrence_id) REFERENCES caldav.recurrence (id)
);

CREATE OR REPLACE PROCEDURE caldav.create_or_update_calendar_file(
    IN _calendar_uid UUID,
    IN _calendar_folder_type caldav.calendar_type,
    IN _etag TIMESTAMP,
    IN _modified_at TIMESTAMP
)
    LANGUAGE plpgsql AS
$$
DECLARE
    _calendar_folder_id BIGINT;
BEGIN
    SELECT id
    INTO _calendar_folder_id
    FROM caldav.calendar_folder
    WHERE type = _calendar_folder_type;
    IF EXISTS (SELECT 1
               FROM caldav.calendar_file
               WHERE uid = _calendar_uid) THEN
        UPDATE caldav.calendar_file
        SET etag        = _etag,
            modified_at = _modified_at
        WHERE uid = _calendar_uid;
    ELSE
        INSERT INTO caldav.calendar_file (uid,
                                          calendar_folder_id,
                                          etag,
                                          created_at,
                                          modified_at)
        VALUES (_calendar_uid,
                _calendar_folder_id,
                _etag,
                now()::timestamp,
                now()::timestamp);
    END IF;
END;
$$;

COMMIT;
