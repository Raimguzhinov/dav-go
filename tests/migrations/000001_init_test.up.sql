BEGIN;

CREATE TABLE IF NOT EXISTS caldav.test_cases
(
    id                      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    calendar_folder_id      BIGINT NOT NULL,
    calendar_file_uid       UUID   NOT NULL,
    event_component_id      BIGINT NOT NULL,
    calendar_property_id    BIGINT NOT NULL,
    custom_property_id      BIGINT,
    recurrence_id           BIGINT,
    recurrence_exception_id BIGINT
);

COMMIT;