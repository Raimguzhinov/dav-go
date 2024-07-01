BEGIN;

CREATE SCHEMA IF NOT EXISTS caldav;

CREATE TYPE caldav.calendar_type AS ENUM ('VEVENT', 'VTODO', 'VJOURNAL');

CREATE TABLE IF NOT EXISTS caldav.calendar_folder
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT                   DEFAULT NULL,
    types       caldav.calendar_type[] DEFAULT ARRAY ['VEVENT']::caldav.calendar_type[],
    max_size    INT                    DEFAULT 4096
);

INSERT INTO caldav.calendar_folder(name, types)
VALUES ('Default Calendar', ARRAY ['VEVENT', 'VTODO', 'VJOURNAL']::caldav.calendar_type[]);

CREATE TABLE IF NOT EXISTS caldav.access
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_folder_id BIGINT UNIQUE REFERENCES caldav.calendar_folder (id) ON DELETE CASCADE,
    user_id            VARCHAR(50) NOT NULL,
    owner              BIT         NOT NULL,
    read               BIT         NOT NULL,
    write              BIT         NOT NULL
);

CREATE TABLE IF NOT EXISTS caldav.calendar_file
(
    uid                UUID PRIMARY KEY,
    calendar_folder_id BIGINT REFERENCES caldav.calendar_folder (id) ON DELETE CASCADE,
    etag               VARCHAR(40) NOT NULL, -- SHA-1 hash encoded in base64
    created_at         TIMESTAMP   NOT NULL,
    modified_at        TIMESTAMP   NOT NULL,
    size               INT         NOT NULL
);

CREATE TABLE IF NOT EXISTS caldav.calendar_property
(
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_file_uid UUID UNIQUE REFERENCES caldav.calendar_file (uid) ON DELETE CASCADE,
    version           VARCHAR(5)   NOT NULL,
    product           VARCHAR(100) NOT NULL,
    scale             VARCHAR(30),
    method            VARCHAR(30)
);

CREATE TABLE IF NOT EXISTS caldav.event_component
(
    id                    BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_file_uid     UUID REFERENCES caldav.calendar_file (uid) ON DELETE CASCADE,
    component_type        BIT       NOT NULL,
    date_timestamp        TIMESTAMP NOT NULL,
    created_at            TIMESTAMP NOT NULL,
    last_modified_at      TIMESTAMP,
    summary               VARCHAR(512),
    description           TEXT,
    url                   TEXT,
    organizer             VARCHAR(255),
    start_date            TIMESTAMP,
    end_date              TIMESTAMP,
    duration              BIGINT,
    all_day               BIT,
    class                 VARCHAR(50),
    location              VARCHAR(255),
    priority              SMALLINT,
    sequence              INT DEFAULT 0,
    status                VARCHAR(50),
    categories            VARCHAR(255),
    event_transparency    BIT,
    todo_completed        DATE,
    todo_percent_complete SMALLINT,
    properties            JSONB,
    UNIQUE (calendar_file_uid, created_at)
);

CREATE TABLE IF NOT EXISTS caldav.attachment
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id BIGINT UNIQUE REFERENCES caldav.event_component (id) ON DELETE CASCADE,
    media_type         VARCHAR(255),
    external_url       TEXT,
    content            BYTEA
);

CREATE TABLE IF NOT EXISTS caldav.attendee
(
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id   BIGINT UNIQUE REFERENCES caldav.event_component (id) ON DELETE CASCADE,
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
    participation_status VARCHAR(15)
);

CREATE TABLE IF NOT EXISTS caldav.alarm
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id BIGINT UNIQUE REFERENCES caldav.event_component (id) ON DELETE CASCADE,
    action             VARCHAR(15) NOT NULL,
    trigger            VARCHAR(15) NOT NULL,
    attachment_id      BIGINT UNIQUE REFERENCES caldav.attachment (id) ON DELETE CASCADE,
    summary            VARCHAR(255),
    description        TEXT,
    attendee_id        BIGINT REFERENCES caldav.attendee (id) ON DELETE CASCADE,
    duration           TIMESTAMP,
    repeat             SMALLINT
);

CREATE TABLE IF NOT EXISTS caldav.recurrence
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id BIGINT UNIQUE REFERENCES caldav.event_component (id) ON DELETE CASCADE,
    interval           INT,
    until              DATE,
    count              INT,
    week_start         SMALLINT,
    by_day             SMALLINT,
    by_month_day       BIGINT,
    by_month           SMALLINT,
    period_day         SMALLINT,
    by_set_pos         INT[],
    this_and_future    BIT
);

CREATE TABLE IF NOT EXISTS caldav.recurrence_exception
(
    event_component_id BIGINT REFERENCES caldav.event_component (id) ON DELETE CASCADE,
    recurrence_id      BIGINT REFERENCES caldav.recurrence (id) ON DELETE CASCADE,
    exception_date     TIMESTAMP NOT NULL,
    deleted_recurrence BIT       NOT NULL,
    PRIMARY KEY (recurrence_id, exception_date)
);

CREATE OR REPLACE PROCEDURE caldav.create_or_update_calendar_file(
    IN p_calendar_uid UUID,
    IN p_calendar_folder_type caldav.calendar_type,
    IN p_calendar_folder_id BIGINT,
    IN p_etag VARCHAR(40),
    IN p_want_etag VARCHAR(40),
    IN p_modified_at TIMESTAMP,
    IN p_size INT,
    IN p_version VARCHAR(5),
    IN p_product VARCHAR(100),
    IN p_if_none_match BOOLEAN DEFAULT FALSE,
    IN p_if_match BOOLEAN DEFAULT FALSE,
    IN p_scale VARCHAR(30) DEFAULT 'GREGORIAN',
    IN p_method VARCHAR(30) DEFAULT NULL
)
    LANGUAGE plpgsql AS
$$
DECLARE
    v_support_folder_id BIGINT;
    v_current_etag      VARCHAR(40);
BEGIN
    SELECT f.id
    INTO
        v_support_folder_id
    FROM caldav.calendar_folder f
    WHERE f.id = p_calendar_folder_id
      AND p_calendar_folder_type = ANY (f.types);

    IF v_support_folder_id IS DISTINCT FROM p_calendar_folder_id THEN
        RAISE EXCEPTION 'Invalid folder type provided for folder: %', p_calendar_folder_id;
    END IF;

    SELECT etag
    INTO
        v_current_etag
    FROM caldav.calendar_file
    WHERE uid = p_calendar_uid;

    IF FOUND THEN
        IF p_if_none_match THEN
            RAISE EXCEPTION 'Precondition failed: If-None-Match header is set and resource exists';
        END IF;

        IF p_if_match AND v_current_etag IS DISTINCT FROM p_want_etag THEN
            RAISE EXCEPTION 'Precondition failed: If-Match header is set and ETag does not match';
        END IF;

        UPDATE
            caldav.calendar_file
        SET etag        = p_etag,
            modified_at = p_modified_at,
            size        = p_size
        WHERE uid = p_calendar_uid;
    ELSE
        IF p_if_match THEN
            RAISE EXCEPTION 'Precondition failed: If-Match header is set and resource does not exist';
        END IF;

        INSERT INTO caldav.calendar_file (uid, calendar_folder_id, etag, created_at, modified_at, size)
        VALUES (p_calendar_uid, p_calendar_folder_id, p_etag, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, p_size);

        INSERT INTO caldav.calendar_property (calendar_file_uid, version, product, scale, method)
        VALUES (p_calendar_uid, p_version, p_product, p_scale, p_method);
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION sequence_update_trigger_fnc()
    RETURNS trigger AS
$$
BEGIN
    SELECT COALESCE(MAX(sequence), 0)
    FROM caldav.event_component
    WHERE calendar_file_uid = NEW.calendar_file_uid
    INTO NEW.sequence;

    NEW.sequence := NEW.sequence + 1;
    RETURN NEW;
END;
$$
    LANGUAGE 'plpgsql';
CREATE TRIGGER sequence_update_trigger
    BEFORE UPDATE
    ON caldav.event_component
    FOR EACH ROW
    WHEN ( OLD.last_modified_at IS DISTINCT FROM NEW.last_modified_at )
EXECUTE PROCEDURE sequence_update_trigger_fnc();

CREATE OR REPLACE FUNCTION recurrence_changed_update_trigger_fnc()
    RETURNS trigger AS
$$
DECLARE
    v_count_recurrence INT;
BEGIN
    SELECT count(*)
    FROM caldav.recurrence
    WHERE recurrence.event_component_id = (SELECT id
                                           FROM caldav.event_component
                                           WHERE event_component.calendar_file_uid =
                                                 (SELECT calendar_file_uid
                                                  FROM caldav.event_component
                                                  WHERE id = NEW.event_component_id)
                                           ORDER BY id
                                           LIMIT 1)
    INTO v_count_recurrence;

    IF v_count_recurrence = 0 THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$
    LANGUAGE 'plpgsql';
CREATE TRIGGER recurrence_ex_update_trigger
    BEFORE INSERT OR UPDATE
    ON caldav.recurrence_exception
    FOR EACH ROW
EXECUTE PROCEDURE recurrence_changed_update_trigger_fnc();

COMMIT;