CREATE TYPE IF NOT EXISTS primary_persons_mode_enum AS ENUM ('FIXED', 'UNLIMITED');
CREATE TYPE IF NOT EXISTS event_date_precision_enum AS ENUM ('DAY', 'MONTH', 'YEAR');
CREATE TYPE IF NOT EXISTS event_date_bound_enum AS ENUM ('EXACT', 'NOT_BEFORE', 'NOT_AFTER');

CREATE TABLE IF NOT EXISTS event_types
(
    id UUID PRIMARY KEY,
    owner_user_id INT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    name TEXT NOT NULL,
    primary_persons_mode primary_persons_mode_enum NOT NULL,
    primary_persons_count INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_event_types_owner_scope
        CHECK ((is_system = TRUE AND owner_user_id = 0) OR (is_system = FALSE AND owner_user_id > 0)),
    CONSTRAINT chk_event_types_primary_count
        CHECK (
            (primary_persons_mode = 'FIXED' AND primary_persons_count >= 1)
            OR
            (primary_persons_mode = 'UNLIMITED' AND primary_persons_count = 0)
        )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_event_types_owner_system_name
    ON event_types (owner_user_id, is_system, lower(name));

CREATE TABLE IF NOT EXISTS events
(
    id UUID PRIMARY KEY,
    tree_id UUID NOT NULL,
    event_type_id UUID NOT NULL REFERENCES event_types (id) ON DELETE RESTRICT,
    date_value DATE NOT NULL,
    date_precision event_date_precision_enum NOT NULL,
    date_bound event_date_bound_enum NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_tree_id ON events (tree_id);
CREATE INDEX IF NOT EXISTS idx_events_event_type_id ON events (event_type_id);

CREATE TABLE IF NOT EXISTS event_primary_persons
(
    event_id UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    person_id UUID NOT NULL,
    position INT NOT NULL,
    PRIMARY KEY (event_id, position),
    UNIQUE (event_id, person_id),
    CONSTRAINT chk_event_primary_position CHECK (position >= 1)
);

CREATE INDEX IF NOT EXISTS idx_event_primary_persons_person_id ON event_primary_persons (person_id);

CREATE TABLE IF NOT EXISTS event_additional_persons
(
    event_id UUID NOT NULL REFERENCES events (id) ON DELETE CASCADE,
    person_id UUID NOT NULL,
    position INT NOT NULL,
    PRIMARY KEY (event_id, position),
    UNIQUE (event_id, person_id),
    CONSTRAINT chk_event_additional_position CHECK (position >= 1)
);

CREATE INDEX IF NOT EXISTS idx_event_additional_persons_person_id ON event_additional_persons (person_id);
