BEGIN;

DROP TABLE IF EXISTS caldav.attachment;
DROP TABLE IF EXISTS caldav.attendee;
DROP TABLE IF EXISTS caldav.alarm;
DROP TABLE IF EXISTS caldav.recurrence_exception;
DROP TABLE IF EXISTS caldav.recurrence;
DROP TABLE IF EXISTS caldav.custom_property;
DROP TABLE IF EXISTS caldav.event_component;
DROP TABLE IF EXISTS caldav.calendar_file;
DROP TABLE IF EXISTS caldav.access;
DROP TABLE IF EXISTS caldav.calendar_folder_property;
DROP TABLE IF EXISTS caldav.calendar_folder;

DROP PROCEDURE IF EXISTS caldav.create_or_update_calendar_file;
DROP TYPE IF EXISTS caldav.calendar_type;

DROP SCHEMA IF EXISTS caldav;

COMMIT;
