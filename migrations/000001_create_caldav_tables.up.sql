BEGIN;

CREATE TYPE public.calendar_type AS ENUM ('VEVENT', 'VTODO', 'VJOURNAL');

CREATE TABLE IF NOT EXISTS public.calendar_folder
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(50) NOT NULL,
    type        calendar_type NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS public.calendar_folder_property
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_folder_id BIGINT       NOT NULL,
    name               VARCHAR(255) NOT NULL,
    namespace          VARCHAR(100) NOT NULL,
    prop_value         TEXT         NOT NULL,
    CONSTRAINT fk_calendar_folder FOREIGN KEY (calendar_folder_id) REFERENCES public.calendar_folder (id)
);

INSERT INTO public.calendar_folder (name, type)
VALUES ('calendars', 'VEVENT');
INSERT INTO public.calendar_folder (name, type)
VALUES ('todos', 'VTODO');
INSERT INTO public.calendar_folder (name, type)
VALUES ('journals', 'VJOURNAL');

CREATE TABLE IF NOT EXISTS public.access
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_folder_id BIGINT      NOT NULL,
    user_id            VARCHAR(50) NOT NULL,
    owner              BIT         NOT NULL,
    read               BIT         NOT NULL,
    write              BIT         NOT NULL,
    CONSTRAINT fk_calendar_folder FOREIGN KEY (calendar_folder_id) REFERENCES public.calendar_folder (id)
);

CREATE TABLE IF NOT EXISTS public.calendar_file
(
    uid                UUID PRIMARY KEY,
    calendar_folder_id BIGINT    NOT NULL,
    etag               TIMESTAMP NOT NULL,
    created_at         TIMESTAMP NOT NULL,
    modified_at        TIMESTAMP NOT NULL,
    CONSTRAINT fk_calendar_folder FOREIGN KEY (calendar_folder_id) REFERENCES public.calendar_folder (id)
);

CREATE TABLE IF NOT EXISTS public.custom_property
(
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id         BIGINT       NOT NULL,
    calendar_file_uid UUID         NOT NULL,
    prop_name         VARCHAR(50)  NOT NULL,
    parameter_name    VARCHAR(50),
    value             varchar(512) NOT NULL,
    CONSTRAINT fk_calendar_file FOREIGN KEY (calendar_file_uid) REFERENCES public.calendar_file (uid)
);

CREATE TABLE IF NOT EXISTS public.event_component
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
    CONSTRAINT fk_calendar_file FOREIGN KEY (calendar_file_uid) REFERENCES public.calendar_file (uid)
);

CREATE TABLE IF NOT EXISTS public.attachment
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id BIGINT NOT NULL,
    uid                UUID   NOT NULL,
    media_type         VARCHAR(255),
    external_url       TEXT,
    content            BYTEA,
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES public.event_component (id)
);

CREATE TABLE IF NOT EXISTS public.attendee
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
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES public.event_component (id)
);

CREATE TABLE IF NOT EXISTS public.alarm
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
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES public.event_component (id)
);

CREATE TABLE IF NOT EXISTS public.recurrence
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
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES public.event_component (id)
);

CREATE TABLE IF NOT EXISTS public.recurrence_exception
(
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    recurrence_id  BIGINT NOT NULL,
    uid            UUID   NOT NULL,
    exception_date DATE   NOT NULL,
    timezone_id    VARCHAR(255),
    all_day        BIT    NOT NULL,
    CONSTRAINT fk_recurrence FOREIGN KEY (recurrence_id) REFERENCES public.recurrence (id)
);

CREATE OR REPLACE PROCEDURE public.create_or_update_calendar_file(
    IN _calendar_uid UUID,
    IN _calendar_folder_type public.calendar_type,
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
    FROM public.calendar_folder
    WHERE type = _calendar_folder_type;
    IF EXISTS (SELECT 1
               FROM public.calendar_file
               WHERE uid = _calendar_uid) THEN
        UPDATE public.calendar_file
        SET etag        = _etag,
            modified_at = _modified_at
        WHERE uid = _calendar_uid;
    ELSE
        INSERT INTO public.calendar_file (uid,
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
