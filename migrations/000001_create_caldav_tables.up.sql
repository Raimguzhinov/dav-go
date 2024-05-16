BEGIN;

CREATE TABLE IF NOT EXISTS public.calendar_folder
(
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(50) NOT NULL,
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
    id                 UUID PRIMARY KEY,
    calendar_folder_id BIGINT    NOT NULL,
    etag               TIMESTAMP NOT NULL,
    created_at         TIMESTAMP NOT NULL,
    modified_at        TIMESTAMP NOT NULL,
    CONSTRAINT fk_calendar_folder FOREIGN KEY (calendar_folder_id) REFERENCES public.calendar_folder (id)
);

CREATE TABLE IF NOT EXISTS public.custom_property
(
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id        BIGINT       NOT NULL,
    calendar_file_id UUID         NOT NULL,
    prop_name        VARCHAR(50)  NOT NULL,
    parameter_name   VARCHAR(50),
    value            varchar(512) NOT NULL,
    CONSTRAINT fk_calendar_file FOREIGN KEY (calendar_file_id) REFERENCES public.calendar_file (id)
);

CREATE TABLE IF NOT EXISTS public.event_component
(
    id                         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_file_id           UUID NOT NULL,
    component_type             BIT  NOT NULL,
    date_timestamp             DATE NOT NULL,
    created_at                 DATE,
    last_modified_at           DATE,
    summary                    VARCHAR(512),
    description                TEXT,
    organizer_email            VARCHAR(255),
    organizer_common_name      VARCHAR(50),
    start_date                 DATE,
    start_timezone_id          VARCHAR(255),
    end_date                   DATE,
    end_timezone_id            VARCHAR(255),
    duration                   BIGINT,
    all_day                    BIT,
    class                      VARCHAR(50),
    location                   VARCHAR(255),
    priority                   SMALLINT,
    sequence                   INT,
    status                     VARCHAR(50),
    categories                 VARCHAR(255),
    recur_interval             INT,
    recur_until                DATE,
    recur_count                INT,
    recur_week_start           VARCHAR(50),
    recur_by_day               VARCHAR(50),
    recur_by_month_day         VARCHAR(50),
    recur_by_month             VARCHAR(50),
    recur_by_set_pos           VARCHAR(50),
    recurrence_id_date         DATE,
    recurrence_id_timezone_id  VARCHAR(255),
    recurrence_this_and_future BIT,
    event_transparency         BIT,
    todo_completed             DATE,
    todo_percent_complete      SMALLINT,
    CONSTRAINT fk_calendar_file FOREIGN KEY (calendar_file_id) REFERENCES public.calendar_file (id)
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

CREATE TABLE IF NOT EXISTS public.recurrence_exception
(
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_component_id BIGINT NOT NULL,
    uid                UUID   NOT NULL,
    exception_date     DATE   NOT NULL,
    timezone_id        VARCHAR(255),
    all_day            BIT    NOT NULL,
    CONSTRAINT fk_event_component FOREIGN KEY (event_component_id) REFERENCES public.event_component (id)
);

CREATE OR REPLACE PROCEDURE create_or_update_calendar_file(
    IN _calendar_uid UUID,
    IN _calendar_folder_name VARCHAR(50),
    IN _etag TIMESTAMP,
    IN _modified_at TIMESTAMP
)
    LANGUAGE plpgsql AS
$$
DECLARE
    _calendar_folder_id BIGINT;
BEGIN
    -- Получим идентификатор папки календаря по ее имени
    SELECT id
    INTO _calendar_folder_id
    FROM calendar_folder
    WHERE name = _calendar_folder_name;
    -- Попробуем найти существующий файл календаря
    IF EXISTS (SELECT 1
               FROM calendar_file
               WHERE id = _calendar_uid) THEN
        -- Если файл существует, обновим поля etag и modified_at
        UPDATE calendar_file
        SET etag        = _etag,
            modified_at = _modified_at
        WHERE id = _calendar_uid;
    ELSE
        -- Иначе вставим новый файл календаря
        INSERT INTO calendar_file (id,
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

INSERT INTO public.calendar_folder (name, description)
VALUES ('VEVENT', 'cals');
INSERT INTO public.calendar_folder (name, description)
VALUES ('VTODO', 'todos');

COMMIT;
