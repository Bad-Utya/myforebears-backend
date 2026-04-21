DROP TABLE IF EXISTS event_additional_persons;
DROP TABLE IF EXISTS event_primary_persons;
DROP TABLE IF EXISTS events;
DROP INDEX IF EXISTS uq_event_types_owner_system_name;
DROP TABLE IF EXISTS event_types;

DROP TYPE IF EXISTS event_date_bound_enum;
DROP TYPE IF EXISTS event_date_precision_enum;
DROP TYPE IF EXISTS primary_persons_mode_enum;
