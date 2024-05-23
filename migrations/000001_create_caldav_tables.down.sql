BEGIN;

DROP TABLE IF EXISTS public.attachment;
DROP TABLE IF EXISTS public.attendee;
DROP TABLE IF EXISTS public.alarm;
DROP TABLE IF EXISTS public.recurrence_exception;
DROP TABLE IF EXISTS public.recurrence;
DROP TABLE IF EXISTS public.custom_property;
DROP TABLE IF EXISTS public.event_component;
DROP TABLE IF EXISTS public.calendar_file;
DROP TABLE IF EXISTS public.access;
DROP TABLE IF EXISTS public.calendar_folder_property;
DROP TABLE IF EXISTS public.calendar_folder;

DROP TYPE IF EXISTS public.calendar_type;
DROP PROCEDURE IF EXISTS create_or_update_calendar_file;

COMMIT;
