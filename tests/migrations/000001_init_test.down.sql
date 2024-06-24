BEGIN;

-- DELETE FROM caldav.attachment WHERE id IN ();
-- DELETE FROM caldav.attendee WHERE id IN ();
-- DELETE FROM caldav.alarm WHERE id IN ();

DELETE FROM caldav.recurrence_exception WHERE id IN (SELECT recurrence_exception_id FROM caldav.test_cases);
DELETE FROM caldav.recurrence WHERE id IN (SELECT recurrence_id FROM caldav.test_cases);
DELETE FROM caldav.custom_property WHERE id IN (SELECT custom_property_id FROM caldav.test_cases);
DELETE FROM caldav.calendar_property WHERE id IN (SELECT calendar_property_id FROM caldav.test_cases);
DELETE FROM caldav.event_component WHERE id IN (SELECT event_component_id FROM caldav.test_cases);
DELETE FROM caldav.calendar_file WHERE uid IN (SELECT calendar_file_uid FROM caldav.test_cases);
-- DELETE FROM caldav.access WHERE id IN ();
-- DELETE FROM caldav.calendar_folder WHERE id IN (SELECT calendar_folder_id FROM caldav.test_cases);

DROP TABLE IF EXISTS caldav.test_cases;

COMMIT;